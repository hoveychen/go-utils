package goutils

import "testing"

func TestCompileRegexp(t *testing.T) {
	r1, err := CompileRegexp("abc")
	if err != nil {
		t.Error(err)
	}
	r2, err := CompileRegexp("abc")
	if err != nil {
		t.Error(err)
	}
	if r1 != r2 {
		t.Error("Regexp not cached")
	}
}

func TestFindNamedStringSubMatch(t *testing.T) {
	cases := []struct {
		Pattern string
		Example string
		Expect  string
	}{
		{
			`https?://.*\.(taobao|tmall)\.com/.*\?.*id=(?P<key>\d+)`,
			`https://item.taobao.com/item.htm?spm=a219r.lm874.14.24.2f436dbctbRumF&id=541127888738&ns=1&abbucket=1#detail`,
			"541127888738",
		},
		{
			`https?://.*\.(taobao|tmall)\.com/.*\?.*id=(?P<key>\d+)`,
			`https://detail.tmall.com/item.htm?id=21860619369&spm=a310p.7395725.1998038907.5.63125564IhHllJ`,
			"21860619369",
		},
		{
			`https?://.*\.brooksbrothers.com/.*\/(?P<key>[A-Za-z0-9]+),default`,
			`http://www.brooksbrothers.com/Non-Iron-Milano-Fit-Glen-Plaid-Sport-Shirt/MG01910,default,pd.html?dwvar_MG01910_Color=MDGR&contentpos=4&cgid=men-fall-sale`,
			"MG01910",
		},
		{
			`https?://.*\.brooksbrothers.com/.*\/(?P<key>[A-Za-z0-9]+),default`,
			`http://www.brooksbrothers.com/Leather-and-Hammered-Gold-Bracelet/SA00001,default,pd.html`,
			"SA00001",
		},
		{
			`https?://.*\.amazon.com/.*dp/(?P<key>[^/\?]+)`,
			`https://www.amazon.com/Stephen-Joseph-Wipeable-Bib-Butterfly/dp/B00N1OFS4U/ref=sr_1_27_s_it?m=ATVPDKIKX0DER&s=baby-products&ie=UTF8&qid=1471860214&sr=1-27&keywords=Stephen+Joseph&refinements=p_85%3A2470955011%2Cp_89%3AStephen+Joseph%2Cp_6%3AATVPDKIKX0DER`,
			"B00N1OFS4U",
		},
		{
			`https?://.*\.amazon.com/.*dp/(?P<key>[^/\?]+)`,
			`http://www.amazon.com/BCBGeneration-Womens-BG-Estelle-Dress-Sandal/dp/B010KA2K8Y/ref=sr_1_11?s=apparel&ie=UTF8&qid=1465021804&sr=1-11&nodeID=679337011&keywords=bcbgeneration+shoes&refinements=p_85%3A2470955011`,
			"B010KA2K8Y",
		},
	}
	for _, c := range cases {
		re, err := CompileRegexp(c.Pattern)
		if err != nil {
			t.Errorf("%s: %v", c.Pattern, err)
			continue
		}
		ret := re.FindNamedStringSubmatch(c.Example)
		if ret == nil {
			t.Errorf("Failed to match: %s", c.Example)
			continue
		}
		if len(ret) == 0 {
			t.Errorf("Matched no named result: %s", c.Example)
			continue
		}
		actual := ret["key"]
		if actual != c.Expect {
			t.Errorf("Extracted wrong result: %s", actual)
			continue
		}
	}
}
