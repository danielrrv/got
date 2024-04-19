package internal_test

import (
	"bytes"
	"testing"

	internal "github.com/danielrrv/got/internal"
)

func TestConfiguration(t *testing.T) {
	t.Run("Unmarshal", func(t *testing.T) {
		var user = internal.UserConfig{
			Name:  "Daniel Rodriguez",
			Email: "drodrigo678@gmail.com",
			Owner: true,
		}
		var config = internal.GotConfig{
			User:   user,
			Bare:   true,
			Branch: "master",
			MaxCache: 10000,
		}
		var ret bytes.Buffer
		internal.Marshal(config, &ret)
		var otherConfig internal.GotConfig
		internal.Unmarshal(ret.Bytes(), &otherConfig)
		if otherConfig.Bare != config.Bare{
			t.Errorf("Expected to be equal")
		}
		if otherConfig.User.Email != config.User.Email{
			t.Errorf("Expected to be equal")
		}
		if otherConfig.MaxCache != config.MaxCache{
			t.Errorf("Expected to be equal")
		}
	})
}
