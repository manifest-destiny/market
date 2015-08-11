package market

import (
	"fmt"
	"reflect"
	"testing"
)

func init() {
	basicMarketScenarios = []*marketTestData{
		{
			title:    "Seller offers less than market price; market offers seller price",
			demand:   1,
			maxPrice: 2,
			balance:  10,
			goodType: Meat,
			bought:   0,
			sellerOffers: []sellerOfferData{
				sellerOfferData{
					price:    1,
					goodType: Meat,
				},
			},
			marketCounters: []marketCounterData{
				marketCounterData{
					offerPrice: 1,
					getsBought: false,
				},
			},
		},
		{
			title:    "Seller offers greater than market price; market counter offers lower than seller price",
			demand:   1,
			maxPrice: 1,
			balance:  10,
			goodType: Meat,
			bought:   0,
			sellerOffers: []sellerOfferData{
				sellerOfferData{
					price:    2,
					goodType: Meat,
				},
			},
			marketCounters: []marketCounterData{
				marketCounterData{
					offerPrice: 1,
					getsBought: true,
				},
			},
		},
		{
			title:    "Seller offers with wrong good type; market returns error",
			demand:   1,
			maxPrice: 1,
			balance:  10,
			goodType: Meat,
			bought:   0,
			sellerOffers: []sellerOfferData{
				sellerOfferData{
					price:    1,
					goodType: Fur,
					expectedError: expectedError{
						isError:   true,
						errorType: "*market.invalidGoodTypeErr",
					},
				},
			},
		},
		{
			title:    "Sellers offers; market second offer is demand error",
			demand:   1,
			maxPrice: 1,
			balance:  10,
			goodType: Meat,
			bought:   0,
			sellerOffers: []sellerOfferData{
				sellerOfferData{
					price:    1,
					goodType: Meat,
				},
				sellerOfferData{
					price:    1,
					goodType: Meat,
				},
			},
			marketCounters: []marketCounterData{
				marketCounterData{
					offerPrice: 1,
					getsBought: true,
				},
				marketCounterData{
					expectedError: expectedError{
						isError:   true,
						errorType: "*market.marketDemandSatisfiedErr",
					},
				},
			},
		},
		{
			title:    "No seller offers is made; market returns error",
			demand:   1,
			maxPrice: 2,
			balance:  10,
			goodType: Meat,
			bought:   0,
			marketCounters: []marketCounterData{
				marketCounterData{
					expectedError: expectedError{
						isError:   true,
						errorType: "*market.noSellerOffersErr",
					},
				},
			},
		},
	}
}

type expectedError struct {
	isError   bool
	errorText string
	errorType string
}

type sellerOfferData struct {
	expectedError
	price    int
	goodType int
}

type marketCounterData struct {
	expectedError
	offerPrice  int
	sellerOffer sellerOfferData
	getsBought  bool
}

type marketTestData struct {
	title          string
	demand         int
	maxPrice       int
	balance        int
	goodType       int
	bought         int
	sellerOffers   []sellerOfferData
	marketCounters []marketCounterData
}

func ExampleMarketConstruction() {
	d := 3
	p := 2
	b := Money(3)
	m := NewBasicMarket(Fur, p, d, b)
	fmt.Println(m.goodType)
	fmt.Println(m.demand)
	fmt.Println(m.maxOffer)
	fmt.Println(m.account.Balance())
	// Output:
	// 1
	// 3
	// 2
	// 3
}

func TestAccountBalance(t *testing.T) {
	a := NewSellerAccount(10)
	if a.Balance() != 10 {
		t.Error("expected", 10, "got", a.Balance())
	}
}

func TestAccountDeposit(t *testing.T) {
	a := NewSellerAccount(0)
	if a.Balance() != 0 {
		t.Error("expected", 0, "got", a.Balance())
	}
	a.Deposit(5)
	if a.Balance() != 5 {
		t.Error("expected", 5, "got", a.Balance())
	}
}

func TestAccountWithdraw(t *testing.T) {
	a := NewSellerAccount(1)
	a.Withdraw(1)
	if a.Balance() != 0 {
		t.Error("expected balance of", 0, "got", a.Balance())
	}
}

func TestAccountWithdrawError(t *testing.T) {
	a := NewSellerAccount(0)
	err := a.Withdraw(1)
	if err == nil {
		t.Error("expected account withdraw error")
	}
}

func createSellerOffer(goodType, price int) *SellerOffer {
	goods := make([]*Good, 1)
	goods[0] = &Good{
		Type: GoodType(goodType),
	}
	return &SellerOffer{
		goods: goods,
		price: Money(price),
	}
}

var basicMarketScenarios []*marketTestData

func TestMarketCounters(t *testing.T) {
	for _, scenario := range basicMarketScenarios {
		market := NewBasicMarket(GoodType(scenario.goodType), scenario.maxPrice, scenario.demand, Money(scenario.balance))
		for _, sOfferScenario := range scenario.sellerOffers {
			err := market.AddOffer(createSellerOffer(sOfferScenario.goodType, sOfferScenario.price))
			handleExpectedErrors(scenario, sOfferScenario.expectedError, err, t)
		}
		for i, mOfferScenario := range scenario.marketCounters {
			marketCounter, err := market.ConsiderOffers()
			handleExpectedErrors(scenario, mOfferScenario.expectedError, err, t)
			if err == nil {
				if Money(mOfferScenario.offerPrice) != marketCounter.price {
					t.Errorf("Expected price of market offer %d to be %d; got %d for \"%s\"", i+1, mOfferScenario.offerPrice, marketCounter.price, scenario.title)
				}
				if mOfferScenario.getsBought {
					err := market.TransactOffer(marketCounter)
					if err != nil {
						t.Error(err.Error())
					}
				}
			}
		}
	}
}

func handleExpectedErrors(test *marketTestData, expected expectedError, err error, t *testing.T) {
	if err == nil && expected.isError == true {
		t.Errorf("Expected an error for \"%s\"", test.title)
	}
	if err != nil && expected.errorType != "" {
		if reflect.TypeOf(err).String() != expected.errorType {
			t.Errorf("Expected error type %s for \"%s\"", reflect.TypeOf(err).String(), test.title)
		}
	}
	if err != nil && expected.errorText != "" {
		if err.Error() != expected.errorText {
			t.Errorf("Expected error \"%s\" for \"%s\"; got \"%s\"", expected.errorText, test.title, err.Error())
		}
	}
	// Catch remaining
	if err != nil && expected.isError == false {
		t.Errorf("Did not expect an error for \"%s\"", test.title)
	}
}
