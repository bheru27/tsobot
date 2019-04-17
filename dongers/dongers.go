package dongers

import "math/rand"

var emoticons map[string][]string

func init() {
	emoticons = map[string][]string{
		"anger":     Anger,
		"disgust":   Disgust,
		"fear":      Fear,
		"happiness": Happiness,
		"neutral":   Neutral,
		"sadness":   Sadness,
		"surprise":  Surprise,
		"panic":     Panic,
	}
}

func Raise(emotion string) (emoticon string) {
	slice, ok := emoticons[emotion]
	if !ok {
		slice = Panic
	}
	return slice[rand.Intn(len(slice))]
}
