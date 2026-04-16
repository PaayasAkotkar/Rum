// Package main showcase the framework
// note: make sure to pass the minIO id of access id and pass or access key in the compose.yml
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	e "rum/app/search-manager"
	"sync"
	"time"
)

var wg sync.WaitGroup

const (
	testAddr     = "localhost:19530"
	denseBucket  = "testBucket"
	hybridBucket = "htestBucket"
)

func main() {
	playRum()
}

func dummyObjects() []e.Object {
	log.SetFlags(log.Lshortfile)

	embedding := make([]float32, e.DefaultDim)
	for i := range embedding {
		embedding[i] = float32(i) * 0.001 // non-zero so MMR dot product is meaningful
	}

	tx := map[int]string{
		1: "Who is the GOAT of chess?",
		2: "What is the best opening for beginners?",
		3: "How do I improve my middle-game strategy?",
	}

	m := make([]e.Object, len(tx))
	for i := 0; i < len(tx); i++ {
		var src e.Object
		src.ID = fmt.Sprintf("ai-replay-%d", i+1)
		src.Branch = "ai-replay-branch"
		src.Text = tx[i+1]
		src.CreatedAt = time.Now()
		src.DocType = "text"
		src.Embedding = embedding

		// Create a structured AI response (Resp)
		// Note: Resp and newDefaultResp are defined in test_main.go (same package)
		resp := newDefaultResp("Rosman")
		*resp.Information.Title = "Grandmaster Analysis"
		*resp.Information.Desc = fmt.Sprintf("Hello Rosman! For your query '%s', it's widely considered that...", tx[i+1])

		respJSON, _ := json.Marshal(resp)
		src.Response = string(respJSON)

		m[i] = src
	}
	return m
}

func playSearch() {

	addr := "localhost:19530"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m, err := e.NewSearch(ctx, addr, "", "", "test", e.DefaultDim)
	if err != nil {
		log.Fatal(err)
	}

	if !m.IsConnected() {
		log.Fatal("milvus not reachable — Fatalping")
	}

	if err := m.Wipe(); err != nil {
		log.Fatal(err)
	}
	log.Println("clean state ensured :)")

	rep := m.Push(denseBucket, testObjects())
	if !rep.IsSucced() {
		log.Fatal("push failed:", rep.Report())
	}
	log.Println(rep.Report())

	candidates, err := m.Pull(denseBucket, "text")
	if err != nil {
		log.Fatalf("pull error: %v", err)
	}
	if len(candidates) == 0 {
		log.Fatal("pull returned zero objects")
	}
	for _, c := range candidates {
		log.Printf("pulled: id=%s branch=%s createdAt=%s", c.Obj.ID, c.Obj.Branch, c.Obj.CreatedAt)
	}
	resp, err := m.Search(denseBucket, testEmbedding(), nil, 4)
	if err != nil {
		log.Fatalf("search error: %v", err)
	}
	if len(resp.Searches) == 0 {
		log.Fatal("search returned zero results")
	}
	for _, r := range resp.Searches {
		log.Printf("search hit: id=%s score=%.4f rawScore=%.4f branch=%s",
			r.Obj.ID, r.Score, r.RawScore, r.Obj.Branch)
	}
	branches := []string{"github:commit:a-1", "github:commit:a-3"}
	resp, err = m.Search(denseBucket, testEmbedding(), branches, 4)
	if err != nil {
		log.Fatalf("search error: %v", err)
	}
	// every result must belong to one of the filtered branches
	for _, r := range resp.Searches {
		found := false
		for _, b := range branches {
			if r.Obj.Branch == b {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("result branch %s not in filter list", r.Obj.Branch)
		}
		log.Printf("filtered hit: id=%s branch=%s score=%.4f", r.Obj.ID, r.Obj.Branch, r.Score)
	}
	resp, err = m.Search(denseBucket, testEmbedding(), nil, 4)
	if err != nil {
		log.Fatalf("search error: %v", err)
	}
	if resp.BestMatch == nil {
		log.Fatal("BestMatch is nil")
	}
	log.Printf("best match: id=%s text=%q score=%.4f", resp.BestMatch.Obj.ID, resp.BestMatch.Obj.Text, resp.BestMatch.Score)
	if err := m.Wipe(); err != nil {
		log.Fatal(err)
	}
	log.Println("dense wiped")
}

func playHybridSearch() {
	addr := "localhost:19530"

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	m, err := e.NewHybridSearch(ctx, addr, "", "", "htest", e.DefaultDim)
	if err != nil {
		log.Println(err)
	}
	if !m.IsConnected() {
		log.Println("unsucceed :(")
		return
	}

	if !m.IsConnected() {
		log.Fatal("milvus not reachable — Fatalping")
	}

	if err := m.Wipe(); err != nil {
		log.Fatal(err)
	}
	log.Println("clean state ensured :)")

	rep := m.Push(hybridBucket, testObjects())
	if !rep.IsSucced() {
		log.Fatalf("push failed: %s", rep.Report())
	}
	log.Println(rep.Report())
	candidates, err := m.Pull(hybridBucket, "text")
	if err != nil {
		log.Fatalf("pull error: %v", err)
	}
	if len(candidates) == 0 {
		log.Fatal("pull returned zero objects")
	}
	for _, c := range candidates {
		log.Printf("pulled: id=%s text=%q", c.Obj.ID, c.Obj.Text)
	}
	resp, err := m.SearchHybrid(hybridBucket, testEmbedding(), "linux golang database", nil, 4)
	if err != nil {
		log.Fatalf("hybrid search error: %v", err)
	}
	if len(resp.Searches) == 0 {
		log.Fatal("hybrid search returned zero results")
	}
	for _, r := range resp.Searches {
		log.Printf("hybrid hit: id=%s text=%q score=%.4f rawScore=%.4f",
			r.Obj.ID, r.Obj.Text, r.Score, r.RawScore)
	}
	branches := []string{"github:commit:a-6", "github:commit:a-10"}
	resp, err = m.SearchHybrid(hybridBucket, testEmbedding(), "golang database", branches, 4)
	if err != nil {
		log.Fatalf("hybrid search error: %v", err)
	}
	for _, r := range resp.Searches {
		found := false
		for _, b := range branches {
			if r.Obj.Branch == b {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("result branch %q not in filter list", r.Obj.Branch)
		}
		log.Printf("filtered hybrid hit: id=%s branch=%s text=%q score=%.4f",
			r.Obj.ID, r.Obj.Branch, r.Obj.Text, r.Score)
	}
	resp, err = m.SearchHybrid(hybridBucket, testEmbedding(), "coffee earth mars", nil, 4)
	if err != nil {
		log.Fatalf("hybrid search error: %v", err)
	}
	if resp.BestMatch == nil {
		log.Fatal("BestMatch is nil")
	}
	log.Printf("best match: id=%s text=%q score=%.4f", resp.BestMatch.Obj.ID, resp.BestMatch.Obj.Text, resp.BestMatch.Score)
	if err := m.Wipe(); err != nil {
		log.Fatal(err)
	}
	log.Println("dense wiped")
}

func testEmbedding() []float32 {
	em := make([]float32, e.DefaultDim)
	for i := range em {
		em[i] = float32(i) * 0.001
	}
	return em
}

func testObjects() []e.Object {
	embedding := testEmbedding()
	texts := map[int]string{
		1:  "hey there",
		2:  "why sky is blue",
		3:  "why linux so famous",
		4:  "my battery is at 100% how do i unplug",
		5:  "any 5 green flower",
		6:  "how to implement a linked list in golang",
		7:  "what is the distance between earth and mars",
		8:  "best practices for restful api design",
		9:  "how to make a perfect cup of coffee at home",
		10: "difference between relational and vector databases",
	}
	objs := make([]e.Object, len(texts))
	for i := range len(texts) {
		objs[i] = e.Object{
			ID:        fmt.Sprintf("obj-%d", i+1),
			Branch:    fmt.Sprintf("github:commit:a-%d", i+1),
			Text:      texts[i+1],
			DocType:   "text",
			Response:  "fdsfsfsd",
			Embedding: embedding,
			CreatedAt: time.Now(),
		}
	}
	return objs
}
