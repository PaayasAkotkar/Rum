// Package rum ...
// flow -> implement the serach-engine easily pass the profile name -> get the bucket -> run the serach -> pass the res
package rum

import (
	"context"
)

// RumStore the main thing you use
type RumStore[In, Out any] struct {
	profile *IProfile[In, Out]
	// search  *searchmanager.SearchManager
	ctx context.Context
}

func NewRumStore[In, Out any](ctx context.Context) *RumStore[In, Out] {
	return &RumStore[In, Out]{
		profile: NewProfile[In, Out](),
		// search:  search,
		ctx: ctx,
	}
}

// get funcs

// // Search embeds the text then searches milvus using the kit's config
// func (r *RumStore[In, Out]) Search(profilename string, embed []float32, text, doctype string) (*searchmanager.SearchResp, error) {
// 	kit, err := r.profile.GetKit(profilename)
// 	if err != nil {
// 		return nil, err
// 	}

// 	a, err := r.search.Pull(kit.Bucket, doctype)
// 	if err != nil {
// 		return nil, err
// 	}
// 	branches := make([]string, 0, len(a))
// 	for _, r := range a {
// 		if r.Obj != nil {
// 			branches = append(branches, r.Obj.Branch)
// 		}
// 	}
// 	if kit.isHybrid {
// 		return r.search.SearchHybrid(kit.Bucket, embed, text, branches, 4)
// 	}
// 	return r.search.Search(kit.Bucket, embed, branches, 4)
// }

func (r *RumStore[In, Out]) GetKit(key string) (*Kit[In, Out], error) {
	return r.profile.GetKit(key)
}

// func (r *RumStore[In, Out]) GetSearch() *searchmanager.SearchManager {
// 	return r.search
// }

// end

// set funcs

// Store embeds and pushes an object into milvus + you handle minIO separately
// func (r *RumStore[In, Out]) Store(key string, embed []float32, id, text, response, branch, docType string) (*searchmanager.IPushReport, error) {
// 	kit, err := r.profile.GetKit(key)
// 	if err != nil {
// 		return nil, err
// 	}

// 	obj := searchmanager.Object{
// 		ID:        id,
// 		Text:      text,
// 		Response:  response,
// 		Branch:    branch,
// 		DocType:   docType,
// 		Embedding: embed,
// 	}

// 	rep := r.search.Push(kit.Bucket, []searchmanager.Object{obj})
// 	return rep, rep.Err
// }

func (r *RumStore[In, Out]) SetProfile(profile *IProfile[In, Out]) {
	r.profile = profile
}

// func (r *RumStore[In, Out]) SetSearch(search *searchmanager.SearchManager) {
// 	r.search = search
// }

// end
