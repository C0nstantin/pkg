// unit tests
package queuer

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConfig_MergeDefaults(t *testing.T) {
	t.Run("merge with default values without env", func(t *testing.T) {
		os.Unsetenv("AMQP_URL")
		c := &Config{
			DSN: "test@test",
		}
		err := c.MergeDefaults()
		if err != nil {
			t.Fatal("Error " + err.Error())
		}
		assert.Equal(t, "test@test", c.DSN)
		assert.Equal(t, "direct", c.ExchangeOptions.Kind)
	})

	t.Run("merge exchange option ", func(t *testing.T) {
		c := &Config{
			ExchangeOptions: ExchangeOptions{
				Kind:       "top",
				Durable:    true,
				AutoDelete: true,
			},
		}
		err := c.MergeDefaults()
		if err != nil {
			return
		}
		assert.Equal(t, "top", c.ExchangeOptions.Kind)
		assert.Equal(t, ExchangeOptionsDefaults.NoWait, c.ExchangeOptions.NoWait)
	})
}
