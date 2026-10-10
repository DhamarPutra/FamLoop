package domain

import (
	"crypto/rand"
	"math/big"
)

// DefaultAvatars berisi kumpulan identitas karakter kartun bawaan yang lucu dan ramah keluarga
var DefaultAvatars = []string{
	"char_lion",
	"char_bear",
	"char_cat",
	"char_fox",
	"char_rabbit",
	"char_panda",
	"char_koala",
	"char_tiger",
	"char_penguin",
	"char_owl",
	"char_dog",
	"char_deer",
}

// GetRandomDefaultAvatar mengembalikan salah satu karakter default secara acak
func GetRandomDefaultAvatar() string {
	if len(DefaultAvatars) == 0 {
		return "char_panda"
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(DefaultAvatars))))
	if err != nil {
		return DefaultAvatars[0]
	}
	return DefaultAvatars[n.Int64()]
}
