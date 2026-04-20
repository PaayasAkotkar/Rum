// Package searchmanager implements the milvus search engine
package searchmanager

// flow:
// SearchManager a serch engine:
// push          ->  stores in minIO: bucket/branch
// pull          ->  bucket & doctype the reason begin simple is because of query
// search-hybird ->  pulls out the max number of resutls which later turns into best match -> as per required there are sparseReq and denseReq the simple annreq does the job
// search        -> basic search nothing special
