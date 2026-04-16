// Package main...
// note: the currrent structs are kept so taht you can test with the models
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	rumrpc "rum/app/misc/rum"
	"rum/app/rum/client"
	rum "rum/app/rum/server"
	e "rum/app/search-manager"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type ChessCoachMiscEntry struct {
	Key   *string             `json:"key,omitempty"`
	Value *ChessCoachMiscItem `json:"value,omitempty"`
}

type ChessCoachMiscItem struct {
	Title   *string `json:"title,omitempty"`
	Desc    *string `json:"desc,omitempty"`
	CanCopy *bool   `json:"canCopy,omitempty"`
	IsLink  *bool   `json:"isLink,omitempty"`
	Link    *string `json:"link,omitempty"`
	Copy    *string `json:"copy,omitempty"`
}

type ChessCoachPayload struct {
	Status  *int32  `json:"status,omitempty"`
	Message *string `json:"message,omitempty"`
}

type ChessCoachReply struct {
	Year      *string                `json:"year,omitempty"`
	Title     *string                `json:"title,omitempty"`
	Desc      *string                `json:"desc,omitempty"`
	Outro     *string                `json:"outro,omitempty"`
	Link      []*ChessCoachMiscEntry `json:"link,omitempty"`
	CopyItems []*ChessCoachMiscEntry `json:"copyItems,omitempty"`
	MiscItems []*ChessCoachMiscEntry `json:"miscItems,omitempty"`
}

type ChessStudentRequest struct {
	ID    *string `json:"id,omitempty"`
	Name  *string `json:"name,omitempty"`
	Query *string `json:"query,omitempty"`
}

type RespX struct {
	Information  *ChessCoachReply       `json:"information,omitempty"`
	Suggestion   *ChessCoachReply       `json:"suggestion,omitempty"`
	BestPractice *ChessCoachReply       `json:"bestPractice,omitempty"`
	MiscItems    []*ChessCoachMiscEntry `json:"miscItems,omitempty"`
}
type Resp struct {
	Stored []e.Object `json:"stored"`
}

type Req struct {
	ID      *string `json:"id,omitempty"`
	Name    *string `json:"name,omitempty"`
	Query   *string `json:"query,omitempty"`
	Profile string  `json:"profile"`
}

var (
	chessCoachFlow func(ctx context.Context, req Req) (*Resp, error)
	flowOnce       sync.Once
)

const (
	MODEL = "googleai/gemini-2.5-flash" // or anyother model
	API   = ""
)

// playRum demonstrates the example of the rum
func playRum() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	addr := "localhost:9300"
	profName := "test_profile"
	bu := "test_bucket"

	// gen := genkit.Init(ctx,
	// 	genkit.WithPlugins(&googlegenai.GoogleAI{APIKey: API}), // Uses GEMINI_API_KEY from env
	// 	genkit.WithDefaultModel(MODEL),
	// )

	// initGeminiFlow(gen)

	engine, _ := e.NewHybridSearch(ctx, "localhost:19530", "", "", bu, e.DefaultDim)
	store := rum.NewRumStore[Req, *Resp](ctx, engine)
	profile := rum.NewProfile[Req, *Resp]()
	kit := rum.NewKit[Req, *Resp](MODEL)
	kit.SetBucket(bu)
	service := rum.NewService[Req, *Resp](ctx, "gem-ser")
	dispatch := rum.NewDispatcher[Req, *Resp]()

	var register rum.IRegister[Req, *Resp]
	register.Fn = func(ctx context.Context, req Req) (*Resp, error) {
		if req.Query == nil {
			q := "default query"
			req.Query = &q
		}

		objs := dummyObjects()
		var resp = Resp{
			Stored: objs,
		}
		for _, obj := range objs {
			_, err := store.Store(req.Profile, obj.Embedding, obj.ID, obj.Text, obj.Response, obj.Branch, obj.DocType)
			if err != nil {
				return nil, err
			}
		}

		return &resp, nil
	}

	dispatch.Register("coach-reply", register)

	service.SetDispatch(dispatch)

	kit.SetService(map[string]*rum.Service[Req, *Resp]{
		"coach-service": service,
	})

	seq := rum.ISequence[Req]{Name: profName, Rank: 1}
	profile.RegisterProfile(seq, kit)
	store.SetProfile(profile)

	rumx := rum.New(ctx, store)

	var wg sync.WaitGroup
	// note: it is not ideal and recommended to use this func like this
	//       it is kept only for demonstration purpose.
	wg.Add(3)
	go func() {
		defer wg.Done()
		rumx.Serve(ctx, rum.RumServer{
			Network: "tcp",
			Address: addr,
		})
	}()

	go func() {
		defer wg.Done()
		res := rumx.Paper(seq)
		if res.IsReady {
			log.Println("metric: ", res.Metric.JSON())
			log.Println("result: ", string(res.Result))
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 2)
		var req Req
		id := "texs1121"
		name := "testRozman"
		query := "who is the goat of chess?"
		req.ID = &id
		req.Name = &name
		req.Query = &query
		req.Profile = seq.Name
		parcel, err := json.Marshal(req)
		if err != nil {
			log.Println("marshal error", err)
			return
		}
		post := rumrpc.IPost{
			Profile: &rumrpc.ISequence{
				Name:  seq.Name,
				Rank:  int32(seq.Rank),
				Input: parcel,
			},
			Push: true,
		}
		client.POST(addr, []*rumrpc.IPost{&post})

	}()

	wg.Wait()
}

func initGeminiFlow(gen *genkit.Genkit) {

	flowOnce.Do(func() {
		flow := genkit.DefineFlow(gen, "gemini", func(ctx context.Context, in Req) (*Resp, error) {
			fullPrompt := prompt(in, "")
			resp, err := genkit.Generate(ctx, gen, ai.WithPrompt(fullPrompt))

			if err != nil {
				log.Println("generate error", err)
				return nil, err
			}
			var res Resp
			if err := json.Unmarshal([]byte(resp.Text()), &res); err != nil {
				log.Println("unmarshal error", err)
				log.Println("raw response:", resp.Text())
				return nil, err
			}

			return &res, nil
		})
		chessCoachFlow = flow.Run
	})
}

func embed(ctx context.Context, em ai.Embedder, text string) ([]float32, error) {
	res, err := em.Embed(ctx, &ai.EmbedRequest{
		Input: []*ai.Document{
			ai.DocumentFromText(text, nil),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("embed: %w", err)
	}
	if len(res.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return res.Embeddings[0].Embedding, nil
}

func prompt(student Req, chunk string) string {
	var sb strings.Builder
	query := ""
	if student.Query != nil {
		query = *student.Query
	}
	name := ""
	if student.Name != nil {
		name = *student.Name
	}

	sb.WriteString(fmt.Sprintf("You are a professional chess coach speaking to %s.\n\n", name))
	sb.WriteString(fmt.Sprintf("IMPORTANT: Start your response in the 'information.desc' field with: 'Hello %s!'\n\n", name))
	sb.WriteString(fmt.Sprintf("Student's Query: %s\n\n", query))

	if len(chunk) > 0 {
		sb.WriteString(chunk)
	} else {
		sb.WriteString("Use your own chess expertise to answer the student's query.\n\n")
	}

	sb.WriteString("JSON STRUCTURE RULES:\n")
	sb.WriteString("- You MUST return valid JSON matching the exact structure below.\n")
	sb.WriteString("- 'information', 'suggestion', and 'bestPractice' are OBJECTS, not arrays.\n")
	sb.WriteString("- Put all lists of items (books, videos, FENs) into 'miscItems' at the top level or within the objects.\n")
	sb.WriteString("- Use 'suggestion' (singular) as the key, not 'suggestions'.\n\n")

	sb.WriteString("FIELD USAGE:\n")
	sb.WriteString("- 'information': General intro and overview.\n")
	sb.WriteString("- 'suggestion': Strategic advice for the specific position/query.\n")
	sb.WriteString("- 'bestPractice': General chess principles applicable here.\n")
	sb.WriteString("- 'miscItems': A list of key-value pairs for tools, resources, and copyable text.\n\n")

	sb.WriteString("MANDATORY JSON TEMPLATE:\n")
	sb.WriteString(`{
  "information": {
    "title": "...",
    "desc": "Hello [Name]! ...",
    "year": "2025"
  },
  "suggestion": {
     "title": "...",
     "desc": "..."
  },
  "bestPractice": {
     "title": "...",
     "desc": "..."
  },
  "miscItems": [
		{ "key": "book1", "value": { "title": "📚 My System", "desc": "Great strategy book...", "canCopy": false, "isLink": false } },
		{ "key": "vid1", "value": { "title": "📺 Guide Video", "desc": "Detailed analysis of master games.", "canCopy": false, "isLink": true, "link": "https://www.youtube.com/watch?v=FULL_ID_HERE" } },
		{"key": "Best Games", "value": { "title": "📺 Best Chess Games of the decade 2019", "desc": "here are some of the best chess game fen you can copy and use and anaylsis.", "canCopy": true,"isLink":false,"copy":"rn1q1rk1/pp2b1pp/2p2n2/3p1pB1/3P4/1QP2N2/PP1N1PPP/R4RK1 b - - 1 11" } }
	                                 ]
}`)
	sb.WriteString("\n\n")
	sb.WriteString("SPECIAL HANDLING:\n")
	sb.WriteString("- For FEN/PGN: Set 'canCopy': true and put the code in 'copy'.\n")
	sb.WriteString("- For Links/Videos: Set 'isLink': true and put the URL in 'link'.\n")
	sb.WriteString("- You MUST include at least one YouTube video link in 'miscItems'.\n\n")

	sb.WriteString("CRITICAL: Return ONLY raw JSON. No markdown code blocks. Ensure 'suggestion' and 'bestPractice' are single objects, NOT arrays.\n")
	return sb.String()
}

func newDefaultResp(studentName string) *RespX {
	title := "Chess Coaching Advice"
	intro := fmt.Sprintf("Hello %s! I see you're interested in improving your chess game. While my AI link is buffering, here's some fundamental advice.", studentName)
	year := "2025"

	trueVal := true
	falseVal := false

	return &RespX{
		Information: &ChessCoachReply{
			Title: &title,
			Desc:  &intro,
			Year:  &year,
		},
		Suggestion: &ChessCoachReply{
			Title: ptrString("Focus on the Center"),
			Desc:  ptrString("Always try to control the four central squares (e4, d4, e5, d5) to maximize your piece activity."),
		},
		BestPractice: &ChessCoachReply{
			Title: ptrString("King Safety"),
			Desc:  ptrString("Don't forget to castle early! Your king is much safer behind a wall of pawns."),
		},
		MiscItems: []*ChessCoachMiscEntry{
			{
				Key: ptrString("vid_default"),
				Value: &ChessCoachMiscItem{
					Title:   ptrString("📺 Chess Fundamentals"),
					Desc:    ptrString("A great video on opening principles."),
					CanCopy: &falseVal,
					IsLink:  &trueVal,
					Link:    ptrString("https://www.youtube.com/watch?v=msy79O_nL5Y"),
				},
			},
		},
	}
}

func ptrString(s string) *string {
	return &s
}
