package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_apolloURLHelper(t *testing.T) {
	url := "https://old.reddit.com/r/Games/comments/k4gz5k/yakuza_like_a_dragon_has_been_out_for_a_while_now/"
	want := "apollo://old.reddit.com/r/Games/comments/k4gz5k/yakuza_like_a_dragon_has_been_out_for_a_while_now/"
	got := string(apolloURLHelper(url))

	require.Equal(t, want, got)
}
