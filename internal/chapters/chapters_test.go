package chapters

import "testing"

func sample() []Chapter {
	return []Chapter{
		{Index: 0, Title: "Intro", StartMs: 0, EndMs: 1000},
		{Index: 1, Title: "Topic", StartMs: 1000, EndMs: 2000},
		{Index: 2, Title: "Outro", StartMs: 2000, EndMs: 3000},
	}
}

func TestFirstLast(t *testing.T) {
	chs := sample()
	if First(chs).Title != "Intro" {
		t.Errorf("First wrong: %+v", First(chs))
	}
	if Last(chs).Title != "Outro" {
		t.Errorf("Last wrong: %+v", Last(chs))
	}
}

func TestAt(t *testing.T) {
	chs := sample()
	got, ok := At(chs, 1)
	if !ok || got.Title != "Topic" {
		t.Errorf("At(1) wrong: %+v ok=%v", got, ok)
	}
	if _, ok := At(chs, 9); ok {
		t.Error("At(9) should be false")
	}
}
