package laukit

import "github.com/shopspring/decimal"

func SmallToBigEthers(req string, quote int64) decimal.Decimal {
	v, _ := decimal.NewFromString(req)
	return v.Mul(decimal.New(10, 0).Pow(decimal.New(quote, 0)))
}

func BigToSmallEthers(req string, quote int64) decimal.Decimal {
	v, _ := decimal.NewFromString(req)
	return v.DivRound(decimal.New(10, 0).Pow(decimal.New(quote, 0)), int32(quote))
}

func WeiToEther(req string) decimal.Decimal {
	return BigToSmallEthers(req, 18)
}

func EtherToWei(req string) decimal.Decimal {
	return SmallToBigEthers(req, 18)
}
