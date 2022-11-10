package couponpusher

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestKingSoopersDoIt(t *testing.T) {
	assert := assert.New(t)
	logrus.SetLevel(logrus.DebugLevel)
	ksc := NewKingSoopersCoupon()
	defer ksc.Teardown()
	// var err error

	t.Run("Can do it all", func(t *testing.T) {
		//t.Skip("Skipping the clicking for now.")
		assert.NoError(ksc.DoIt())
	})
}

func TestRemoveAllCoupons(t *testing.T) {
	assert := assert.New(t)
	logrus.SetLevel(logrus.DebugLevel)
	ksc := NewKingSoopersCoupon()
	defer ksc.Teardown()
	var err error
	t.Run("Can login", func(t *testing.T) {
		err = ksc.Login()
		assert.NoError(err)
	})
	err = ksc.RemoveAllCoupons()
	assert.NoError(err)
}
