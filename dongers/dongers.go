package dongers

import "math/rand"

var emoticons map[string][]string

func init() {
	emoticons = map[string][]string{
		"Anger":     Anger,
		"Disgust":   Disgust,
		"Fear":      Fear,
		"Happiness": Happiness,
		"Neutral":   Neutral,
		"Sadness":   Sadness,
		"Surprise":  Surprise,
		"Panic":     Panic,
	}
}

func Raise(emotion string) (emoticon string) {
	slice, ok := emoticons[emotion]
	if !ok {
		slice = Panic
	}
	return slice[rand.Intn(len(slice))]
}
