package main

import (
	"github.com/sirupsen/logrus"
	couponpusher "github.com/TopherGopher/OffTheGrid/king_soopers_coupon"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	ksc := couponpusher.NewKingSoopersCoupon()
	defer ksc.Teardown()
	if err := ksc.DoIt(); err != nil {
		panic(err)
	}
}
