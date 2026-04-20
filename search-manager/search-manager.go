package searchmanager

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

const (
	FieldID              = "id"
	FieldText            = "text"
	FieldEmbedding       = "embedding"
	FieldResponse        = "response"
	FieldBranch          = "branch"
	FieldDocType         = "docType"
	FieldCreatedAt       = "createdAt"
	FieldSparseEmbedding = "sparseEmbedding"

	DefaultDim  = 768 // best-for-gemini
	DefaultTopK = 5

	MaxTextLen   = 65535
	MaxIDLen     = 128
	MaxBranchLen = 512
)

// Object is the single unit stored and retrieved from Milvus.
type Object struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Embedding []float32 `json:"embedding"`
	Response  string    `json:"response"`
	Branch    string    `json:"branch"`
	DocType   string    `json:"docType"`
	CreatedAt time.Time `json:"createdAt"`
}

// ISearch is a search result with MMR-reranked score.
type ISearch struct {
	Obj      *Object
	Score    float64 // mmr-reranked
	RawScore float32 // raw ANN score from milvus
}

// ICandidate is an intermediate candidate before MMR reranking.
type ICandidate struct {
	Obj      *Object
	RawScore float32
}

// SearchManager wraps the Milvus client.
type SearchManager struct {
	client *milvusclient.Client
	ctx    context.Context
	schema *entity.Schema
	topK   int
	dim    int
}

// Close releases the underlying gRPC connection.
func (m *SearchManager) Close() error {
	ctx := m.ctx
	return m.client.Close(ctx)
}

// IsConnected returns true if Milvus is reachable.
func (m *SearchManager) IsConnected() bool {
	if m == nil || m.client == nil {
		return false
	}
	v, err := m.client.GetServerVersion(m.ctx, milvusclient.NewGetServerVersionOption())
	if err != nil {
		log.Println(err)
		return false
	}
	log.Printf("connected to milvus %s :)", v)
	return true
}

// InitContext replaces the stored context.
func (m *SearchManager) InitContext(ctx context.Context) {
	m.ctx = ctx
}

// IsEmpty returns true if the bucket collection does not exist.
func (m *SearchManager) IsEmpty(bucket string) (bool, error) {
	has, err := m.client.HasCollection(m.ctx, milvusclient.NewHasCollectionOption(bucket))
	if err != nil {
		return false, err
	}
	return !has, nil
}

// EnsureCollection creates the collection + HNSW index if it does not exist,
// then loads it into memory. Safe to call multiple times.
func (m *SearchManager) EnsureCollection(ctx context.Context, bucket string) error {
	if bucket == "" {
		return fmt.Errorf("collection name is empty")
	}

	has, err := m.client.HasCollection(ctx, milvusclient.NewHasCollectionOption(bucket))
	if err != nil {
		return fmt.Errorf("check collection %q: %w", bucket, err)
	}
	if has {
		log.Println("collection included")
		return m.loadIfNeeded(ctx, bucket)
	}

	// // set collection name to bucket at creation time
	// m.schema.CollectionName = bucket

	if err := m.client.CreateCollection(ctx,
		milvusclient.NewCreateCollectionOption(bucket, m.schema)); err != nil {
		return fmt.Errorf("create collection %q: %w", bucket, err)
	}

	// Create indexes for the collection.
	// Dense search always needs an embedding index.
	idxOpt := milvusclient.NewCreateIndexOption(bucket, FieldEmbedding,
		index.NewHNSWIndex(entity.IP, 16, 200))
	if _, err := m.client.CreateIndex(ctx, idxOpt); err != nil {
		return fmt.Errorf("create index on %q: %w", FieldEmbedding, err)
	}

	// Create a sparse index for hybrid search if the schema has a sparse vector field.
	hasSparse := false
	for _, f := range m.schema.Fields {
		if f.Name == FieldSparseEmbedding {
			hasSparse = true
			break
		}
	}
	if hasSparse {
		sparseIdxOpt := milvusclient.NewCreateIndexOption(bucket, FieldSparseEmbedding,
			index.NewSparseInvertedIndex(entity.BM25, 0.2))
		if _, err := m.client.CreateIndex(ctx, sparseIdxOpt); err != nil {
			return fmt.Errorf("create index on %q: %w", FieldSparseEmbedding, err)
		}
	}

	if err := m.loadIfNeeded(ctx, bucket); err != nil {
		return err
	}

	log.Printf("milvus: collection %q created (dim=%d)", bucket, m.dim)
	return nil
}

// loadIfNeeded loads the collection into memory.
// Milvus requires load before search or query.
// Already-loaded collections are silently skipped.
func (m *SearchManager) loadIfNeeded(ctx context.Context, bucket string) error {
	_, err := m.client.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(bucket))

	if err != nil {
		// not fatal — collection may already be loaded
		log.Printf("loadIfNeeded %q (may already be loaded): %v", bucket, err)
	}
	return nil
}

// Push inserts objects into a dense-search bucket.
// Collection is created and indexed automatically on first call.
// Push inserts objects into a bucket.
// Works for both dense-only and hybrid collections —
// if the schema has sparseEmbedding, Milvus BM25 fills it automatically from text.
func (m *SearchManager) Push(bucket string, objs []Object) *IPushReport {
	rep := &IPushReport{
		Bucket:    bucket,
		Hash:      generateMilvusHash(),
		CreatedAt: time.Now(),
	}
	if len(objs) == 0 {
		rep.Err = fmt.Errorf("no objects to insert")
		return rep
	}
	if len(objs) > 0 {
		rep.Branch = objs[0].Branch
	}

	if err := m.EnsureCollection(m.ctx, bucket); err != nil {
		rep.Err = err
		return rep
	}

	ids, texts, branches, docTypes, responses, ebds, createdAts, err := validateAndExtract(objs, m.dim)
	if err != nil {
		rep.Err = err
		return rep
	}

	if _, err := m.insertAndReport(m.ctx, bucket, ids, texts, branches, docTypes, responses, ebds, createdAts); err != nil {
		rep.Err = err
		return rep
	}

	rep.Total = len(ids)
	rep.Dim = m.dim
	for i, id := range ids {
		rep.Objects = append(rep.Objects,
			fmt.Sprintf("%s/%s/%s", bucket, branches[i], id))
	}
	log.Printf("milvus push: hash=%s bucket=%s branch=%s total=%d",
		rep.Hash, rep.Bucket, rep.Branch, rep.Total)
	return rep
}

// validateAndExtract validates objs and extracts columns — shared by Push and HPush.
func validateAndExtract(objs []Object, dim int) (
	ids, texts, branches, docTypes, responses []string,
	ebds [][]float32,
	createdAts []int64,
	err error,
) {
	ids = make([]string, len(objs))
	texts = make([]string, len(objs))
	branches = make([]string, len(objs))
	docTypes = make([]string, len(objs))
	responses = make([]string, len(objs))
	ebds = make([][]float32, len(objs))
	createdAts = make([]int64, len(objs))

	for i, c := range objs {
		if len(c.ID) > MaxIDLen {
			return nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("id too long at index %d", i)
		}
		if len(c.Branch) > MaxBranchLen {
			return nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("branch too long at index %d", i)
		}
		if len(c.Embedding) == 0 {
			return nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("empty embedding at index %d", i)
		}
		if len(c.Embedding) != dim {
			return nil, nil, nil, nil, nil, nil, nil, fmt.Errorf("embedding dim mismatch at index %d: got %d want %d", i, len(c.Embedding), dim)
		}

		ids[i] = c.ID
		texts[i] = c.Text
		branches[i] = c.Branch
		docTypes[i] = c.DocType
		responses[i] = c.Response
		ebds[i] = c.Embedding
		if c.CreatedAt.IsZero() {
			createdAts[i] = time.Now().Unix()
		} else {
			createdAts[i] = c.CreatedAt.Unix()
		}
	}
	return
}

// insertAndReport inserts columns and returns a human-readable report.
func (m *SearchManager) insertAndReport(
	ctx context.Context,
	bucket string,
	ids, texts, branches, docTypes, responses []string,
	ebds [][]float32,
	createdAts []int64,
) (string, error) {
	cols := []column.Column{
		column.NewColumnVarChar(FieldID, ids),
		column.NewColumnVarChar(FieldText, texts),
		column.NewColumnVarChar(FieldBranch, branches),
		column.NewColumnVarChar(FieldDocType, docTypes),
		column.NewColumnVarChar(FieldResponse, responses),
		column.NewColumnFloatVector(FieldEmbedding, m.dim, ebds),
		column.NewColumnInt64(FieldCreatedAt, createdAts),
	}

	if _, err := m.client.Insert(ctx,
		milvusclient.NewColumnBasedInsertOption(bucket).WithColumns(cols...)); err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- Push Report [%s] ---\n", bucket))
	sb.WriteString(fmt.Sprintf("Total Records: %d\n", len(ids)))
	sb.WriteString("Storume Format: Columnar (LSM-Tree based)\n\n")
	for i := range ids {
		sb.WriteString(fmt.Sprintf("[%d] ID: %s | Type: %s | Branch: %s\n",
			i, ids[i], docTypes[i], branches[i]))
		if responses[i] != "" {
			sb.WriteString(fmt.Sprintf("    Response: %s\n", responses[i]))
		}
		if len(ebds[i]) >= 3 {
			sb.WriteString(fmt.Sprintf("    Embedding (dim %d): [%.4f, %.4f, %.4f...]\n",
				m.dim, ebds[i][0], ebds[i][1], ebds[i][2]))
		}
		sb.WriteString(fmt.Sprintf("    Timestamp: %d\n\n", createdAts[i]))
	}
	sb.WriteString("succeed 🤗\n")
	return sb.String(), nil
}

// Delete removes a single object by id from the collection.
func (m *SearchManager) Delete(bucket, id string) error {
	expr := fmt.Sprintf(`%s == "%s"`, FieldID, id)
	if _, err := m.client.Delete(m.ctx,
		milvusclient.NewDeleteOption(bucket).WithExpr(expr)); err != nil {
		return fmt.Errorf("milvus delete failed: %w", err)
	}
	log.Printf("deleted id=%s from bucket=%s", id, bucket)
	return nil
}

// Pull fetches all objects matching docType from bucket, sorted by CreatedAt.
// Useful for retrieving all objects of a specific type (e.g. "image/png", "text/plain").
func (m *SearchManager) Pull(bucket, docType string) ([]ICandidate, error) {
	// load before query
	if err := m.loadIfNeeded(m.ctx, bucket); err != nil {
		return nil, err
	}

	exp := fmt.Sprintf(`%s == "%s"`, FieldDocType, docType)
	res, err := m.client.Query(m.ctx,
		milvusclient.NewQueryOption(bucket).
			WithFilter(exp).
			WithOutputFields(FieldID, FieldText, FieldEmbedding, FieldResponse, FieldBranch, FieldDocType, FieldCreatedAt))
	if err != nil {
		return nil, err
	}
	if res.ResultCount == 0 {
		return nil, nil // not an error — just empty
	}

	var mpull []ICandidate
	for k := range res.ResultCount {
		var t Object

		idIdx := fieldIndex(res.Fields, FieldID)
		txtIdx := fieldIndex(res.Fields, FieldText)
		ebdIdx := fieldIndex(res.Fields, FieldEmbedding)
		resIdx := fieldIndex(res.Fields, FieldResponse)
		brnIdx := fieldIndex(res.Fields, FieldBranch)
		docIdx := fieldIndex(res.Fields, FieldDocType)
		catIdx := fieldIndex(res.Fields, FieldCreatedAt)

		// skip row if any field is missing
		if idIdx < 0 || txtIdx < 0 || ebdIdx < 0 || resIdx < 0 || brnIdx < 0 || docIdx < 0 || catIdx < 0 {
			continue
		}

		rid, err := res.Fields[idIdx].GetAsString(k)
		if err != nil {
			continue
		}
		rtxt, err := res.Fields[txtIdx].GetAsString(k)
		if err != nil {
			continue
		}
		rebd, err := res.Fields[ebdIdx].Get(k)
		if err != nil {
			continue
		}
		rbrn, err := res.Fields[brnIdx].GetAsString(k)
		if err != nil {
			continue
		}
		rdoc, err := res.Fields[docIdx].GetAsString(k)
		if err != nil {
			continue
		}
		rcat, err := res.Fields[catIdx].Get(k)
		if err != nil {
			continue
		}
		rres, err := res.Fields[resIdx].GetAsString(k)
		if err != nil {
			continue
		}

		t.ID = rid
		t.Text = rtxt
		t.Branch = rbrn
		t.DocType = rdoc
		t.Response = rres

		if fv, ok := rebd.(entity.FloatVector); ok {
			t.Embedding = []float32(fv)
		} else if vec, ok := rebd.([]float32); ok {
			t.Embedding = vec
		} else {
			continue
		}

		if val, ok := rcat.(int64); ok {
			t.CreatedAt = time.Unix(val, 0)
		} else if val32, ok := rcat.(int32); ok {
			t.CreatedAt = time.Unix(int64(val32), 0)
		} else {
			continue
		}

		rscore := float32(0)
		if k < len(res.Scores) {
			rscore = res.Scores[k]
		}
		mpull = append(mpull, ICandidate{Obj: &t, RawScore: rscore})
	}

	sort.Slice(mpull, func(i, j int) bool {
		return mpull[i].Obj.CreatedAt.Before(mpull[j].Obj.CreatedAt)
	})
	return mpull, nil
}

type SearchResp struct {
	BestMatch *ISearch
	Searches  []ISearch
}

func (s *SearchResp) Join(current string) string {
	if s.BestMatch == nil || s.BestMatch.Obj == nil {
		return fmt.Sprintf("### PRESENT QUERY ###\nQUERY: %s\n", current)
	}

	var sb strings.Builder
	prevReq := s.BestMatch.Obj.Text
	prevResp := s.BestMatch.Obj.Response

	sb.WriteString("### PREVIOUS SIMILAR CONTEXT ###\n")
	sb.WriteString(fmt.Sprintf("PAST QUERY: %s\n", prevReq))
	sb.WriteString(fmt.Sprintf("PAST RESPONSE: %s\n", prevResp))
	sb.WriteString("\n---\n\n")
	sb.WriteString("### PRESENT TASK ###\n")
	sb.WriteString(fmt.Sprintf("CURRENT QUERY: %s\n", current))

	return sb.String()
}

// mmr is the shared MMR reranking helper used by both Search and SearchHybrid.
// Reads all fields directly from the ResultSet — no extra round trips.
// MMR balances relevance (rawScore) and diversity (max cosine sim to already selected).
// lambda=0.8 weights relevance heavily — raise toward 1.0 for pure relevance,
// lower toward 0.5 for more diversity.
func (m *SearchManager) mmr(topK int, rs []milvusclient.ResultSet) ([]ISearch, error) {
	if len(rs) == 0 {
		return nil, nil
	}

	candidates := make([]ICandidate, 0)
	for r := range len(rs) {
		for i := 0; i < rs[r].ResultCount; i++ {
			// declare cand inside inner loop — prevents RawScore bleeding across iterations
			var cand ICandidate
			var t Object

			idIdx := fieldIndex(rs[r].Fields, FieldID)
			txtIdx := fieldIndex(rs[r].Fields, FieldText)
			ebdIdx := fieldIndex(rs[r].Fields, FieldEmbedding)
			resIdx := fieldIndex(rs[r].Fields, FieldResponse)
			brnIdx := fieldIndex(rs[r].Fields, FieldBranch)
			docIdx := fieldIndex(rs[r].Fields, FieldDocType)
			catIdx := fieldIndex(rs[r].Fields, FieldCreatedAt)

			// skip row if any field is missing
			if idIdx < 0 || txtIdx < 0 || ebdIdx < 0 || resIdx < 0 || brnIdx < 0 || docIdx < 0 || catIdx < 0 {
				continue
			}

			id, err := rs[r].Fields[idIdx].GetAsString(i)
			if err != nil {
				continue
			}
			txt, err := rs[r].Fields[txtIdx].GetAsString(i)
			if err != nil {
				continue
			}
			ebd, err := rs[r].Fields[ebdIdx].Get(i)
			if err != nil {
				continue
			}
			brn, err := rs[r].Fields[brnIdx].GetAsString(i)
			if err != nil {
				continue
			}
			doc, err := rs[r].Fields[docIdx].GetAsString(i)
			if err != nil {
				continue
			}
			cat, err := rs[r].Fields[catIdx].Get(i)
			if err != nil {
				continue
			}
			rres, err := rs[r].Fields[resIdx].GetAsString(i)
			if err != nil {
				continue
			}

			t.ID = id
			t.Text = txt
			t.Branch = brn
			t.DocType = doc
			t.Response = rres

			if fv, ok := ebd.(entity.FloatVector); ok {
				t.Embedding = []float32(fv)
			} else if vec, ok := ebd.([]float32); ok {
				t.Embedding = vec
			} else {
				continue
			}

			if val, ok := cat.(int64); ok {
				t.CreatedAt = time.Unix(val, 0)
			} else if val32, ok := cat.(int32); ok {
				t.CreatedAt = time.Unix(int64(val32), 0)
			} else {
				continue
			}

			cand.Obj = &t
			cand.RawScore = rs[r].Scores[i]
			candidates = append(candidates, cand)
		}
	}

	if topK > len(candidates) {
		topK = len(candidates)
	}
	if topK == 0 {
		return nil, nil
	}

	const lambda = float32(0.8)
	selected := make([]ISearch, 0, topK)
	remaining := make([]ICandidate, len(candidates))
	copy(remaining, candidates)

	for len(selected) < topK && len(remaining) > 0 {
		bestIdx := -1
		bestMMR := float32(-1e9)

		for i, cand := range remaining {
			maxSim := float32(0)
			for _, sel := range selected {
				if s := dotPrd(cand.Obj.Embedding, sel.Obj.Embedding); s > maxSim {
					maxSim = s
				}
			}
			mmrScore := lambda*cand.RawScore - (1-lambda)*maxSim
			if mmrScore > bestMMR {
				bestMMR = mmrScore
				bestIdx = i
			}
		}

		best := remaining[bestIdx]
		selected = append(selected, ISearch{
			Obj:      best.Obj,
			Score:    float64(bestMMR),
			RawScore: best.RawScore,
		})
		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}

	log.Printf("mmr: candidates=%d selected=%d", len(candidates), len(selected))
	return selected, nil
}

// Wipe drops all collections in the database. Use with caution.
func (m *SearchManager) Wipe() error {
	col, err := m.client.ListCollections(m.ctx, milvusclient.NewListCollectionOption())
	if err != nil {
		return err
	}
	if len(col) == 0 {
		log.Println("nothing to wipe :)")
		return nil
	}
	for _, a := range col {
		log.Println("erasing col:", a)
		if err := m.client.DropCollection(m.ctx, milvusclient.NewDropCollectionOption(a)); err != nil {
			return err
		}
	}
	log.Println("wipe succeed :)")
	return nil
}

// Search performs dense ANN search and returns MMR-reranked results.
// embed: dense vector from your encoder — must be dim=768 (or whatever dim was set).
// maxLim: multiplier on topK to widen the candidate pool before MMR reranking.
func (m *SearchManager) Search(
	bucket string,
	embed []float32,
	branches []string,
	maxLim int,
) (*SearchResp, error) {
	if len(embed) != m.dim {
		return nil, fmt.Errorf("embedding dim mismatch: got %d want %d", len(embed), m.dim)
	}
	if maxLim <= 0 {
		maxLim = 4
	}

	opt := milvusclient.NewSearchOption(bucket, m.topK*maxLim,
		[]entity.Vector{entity.FloatVector(embed)}).
		WithANNSField(FieldEmbedding).
		WithSearchParam("metric_type", "IP").
		WithSearchParam("nprobe", "128").
		WithOutputFields(FieldID, FieldText, FieldEmbedding, FieldResponse, FieldBranch, FieldDocType, FieldCreatedAt)

	if f := branchFilter(branches); f != "" {
		opt = opt.WithFilter(f)
	}

	rs, err := m.client.Search(m.ctx, opt)
	if err != nil {
		return nil, err
	}
	a, err := m.mmr(m.topK, rs)
	return &SearchResp{BestMatch: BestMatch(a), Searches: a}, err
}

// SearchHybrid performs dense ANN + BM25 hybrid search and returns MMR-reranked results.
// embed: dense vector from your encoder — must match dim.
// queryText: raw query string for BM25 — Milvus tokenises it internally.
// maxLim: multiplier on topK to widen the candidate pool before MMR reranking.
func (m *SearchManager) SearchHybrid(
	bucket string,
	embed []float32,
	queryText string,
	branches []string,
	maxLim int,
) (*SearchResp, error) {
	if len(embed) != m.dim {
		return nil, fmt.Errorf("embedding dim mismatch: got %d want %d", len(embed), m.dim)
	}
	if maxLim <= 0 {
		maxLim = 4
	}

	lim := m.topK * maxLim
	filter := branchFilter(branches)

	// dense ANN request
	denseReq := milvusclient.NewAnnRequest(FieldEmbedding, lim,
		entity.FloatVector(embed)).
		WithANNSField(FieldEmbedding).
		WithSearchParam("metric_type", "IP").
		WithSearchParam("nprobe", "16")

	// sparse BM25 request
	sparseReq := milvusclient.NewAnnRequest(FieldSparseEmbedding, lim,
		entity.Text(queryText)).
		WithANNSField(FieldSparseEmbedding).
		WithSearchParam("metric_type", "BM25")

	if filter != "" {
		denseReq = denseReq.WithFilter(filter)
		sparseReq = sparseReq.WithFilter(filter)
	}

	rs, err := m.client.HybridSearch(m.ctx,
		milvusclient.NewHybridSearchOption(bucket, lim, denseReq, sparseReq).
			WithReranker(milvusclient.NewRRFReranker().WithK(60)).
			WithOutputFields(FieldID, FieldText, FieldEmbedding, FieldResponse, FieldBranch, FieldDocType, FieldCreatedAt))
	if err != nil {
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}

	a, err := m.mmr(m.topK, rs)
	return &SearchResp{BestMatch: BestMatch(a), Searches: a}, err
}
