package search_test

// func TestSearch(t *testing.T) {
// 	// create a new searcher
// 	instances := []search.Instance{
// 		{Name: "foo", Env: "dev", Team: "foo"},
// 		{Name: "foobar", Env: "prod", Team: "bar"},
// 		{Name: "foobarbaz", Env: "staging", Team: "baz"},
// 	}

// 	teams := []string{"foob"}

// 	s := search.New(instances, teams)

// 	results, err := s.Search("foo", "team")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	want := []search.Result{
// 		{URL: "foo", Type: "instance"},
// 		{URL: "foob", Type: "team"},
// 		{URL: "foobar", Type: "instance"},
// 		{URL: "foobarbaz", Type: "instance"},
// 	}

// 	opts := cmp.Options{
// 		cmpopts.IgnoreFields(search.Result{}, "Rank"),
// 	}

// 	if !cmp.Equal(results, want, opts) {
// 		t.Errorf("diff -want +got:\n%v", cmp.Diff(want, results, opts))
// 	}
// }
