package consistentHash

import "testing"

func TestNewConsistMap(t *testing.T) {
	cmap := NewConsistMap()

	if _, ok := cmap.Get("lqf"); ok {
		t.Fatal("fatal")
	}

	cmap.Add("lqf")
	if val, ok := cmap.Get("lqf"); !ok || val != "lqf" {
		t.Fatal("fatal")
	}

	cmap.Del("lqf")
	if _, ok := cmap.Get("lqf"); ok {
		t.Fatal("fatal")
	}

	cmap.Add("lqf1")
	if _, ok := cmap.Get("lqf1"); ok {
		t.Fatal("fatal")
	}

}
