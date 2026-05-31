package llm

import "sort"

// Prompt is a curated system+user prompt pair.
type Prompt struct {
	System string
	User   string
}

var presets = map[string]Prompt{
	"conclusion": {
		System: "You are a warm, witty podcast host wrapping up an episode of Connected, a show about Apple and technology hosted by Myke Hurley, Stephen Hackett, and Federico Viticci. Write in the friendly, slightly nerdy voice of the show.",
		User:   "Below is the transcript of the final chapter of an episode. Write a short, heartfelt closing summary (3-4 sentences) that captures the spirit of how the episode ended.",
	},
	"cold-open": {
		System: "You are the announcer for the Connected podcast (Myke Hurley, Stephen Hackett, Federico Viticci).",
		User:   "Below is the transcript of the opening chapter of an episode. Write a punchy 2-sentence cold-open teaser for it.",
	},
	"recap": {
		System: "You summarize tech podcast episodes concisely and accurately.",
		User:   "Summarize the key points from this transcript as 4-6 tight bullet points.",
	},
	"style-federico": {
		System: "You write in the enthusiastic, detail-loving voice of Federico Viticci, who always has a lot of thoughts and signs off with 'Arrivederci.'",
		User:   "Write a short outro for this episode in Federico's voice, ending with 'Arrivederci.'",
	},
	"style-myke": {
		System: "You write in the warm, organized, broadcaster voice of Myke Hurley.",
		User:   "Write a short outro for this episode in Myke's voice.",
	},
	"style-stephen": {
		System: "You write in the dry, history-loving voice of Stephen Hackett.",
		User:   "Write a short outro for this episode in Stephen's voice.",
	},
	"haiku": {
		System: "You distill content into haiku (5-7-5).",
		User:   "Write a single haiku capturing this chapter.",
	},
}

// Preset returns a prompt preset by name.
func Preset(name string) (Prompt, bool) {
	p, ok := presets[name]
	return p, ok
}

// PresetNames lists preset names, sorted.
func PresetNames() []string {
	names := make([]string, 0, len(presets))
	for k := range presets {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
