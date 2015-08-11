// Package market provides market and seller interfaces for manifest
// destiny.
package market

import (
	"errors"
	"sync"
)

const (
	Meat = iota
	Fur
)

type Money uint

// GoodType defines a type of good (eg, meat or fur).
type GoodType int

// Account is an interface for managing a player's (or
// other type's) money.
type Account interface {
	Balance() Money
	Withdraw(m Money) error
	Deposit(m Money)
	Lock()
	Unlock()
}

// SellerAccount is for managing a player's money.
type SellerAccount struct {
	balance Money
	sync.Mutex
}

// NewSellerAccount is a constructor function for SellerAccount.
func NewSellerAccount(balance Money) *SellerAccount {
	return &SellerAccount{balance, sync.Mutex{}}
}

// Balance returns the SellerAccount's balance.
func (a *SellerAccount) Balance() Money {
	return a.balance
}

// Withdraw withdraws money from the SellerAccount.
func (a *SellerAccount) Withdraw(m Money) error {
	var err error
	if m <= a.balance {
		a.balance -= m
	} else {
		err = errors.New("Not enough money in account")
	}
	return err
}

// Deposit deposits money into the SellerAccount.
func (a *SellerAccount) Deposit(m Money) {
	// fmt.Println(a.balance)
	// fmt.Println(&m, m)
	a.balance += m
}

// Market interface for selling goods.
type Market interface {
	// ConsiderOffers returns a counter offer or accepted offer
	// for SellerOffers and moves the offer pointer forward.
	ConsiderOffers() (*MarketCounter, error)
	// TransactOffer handles the transaction for the MarketOffer.
	TransactOffer(c *MarketCounter) error
}

// Good describes some kind of good (eg, meat or fur).
type Good struct {
	Type       GoodType
	Refinement int
}

// SellerOffer is slice of Goods that can be offered to a Market for
// a price.
type SellerOffer struct {
	goods   []*Good
	account *SellerAccount
	price   Money
}

// ByOfferPrice is a sort interface for sorting a slice of SellerOffers
// by price.
type ByOfferPrice []*SellerOffer

func (a ByOfferPrice) Len() int           { return len(a) }
func (a ByOfferPrice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByOfferPrice) Less(i, j int) bool { return a[i].price < a[j].price }

// MarketCounter is an accepted offer or a counter offer for a SellerOffer.
type MarketCounter struct {
	sellerOffer *SellerOffer
	price       Money
}

// Transfer function enables transfers between two account.
func Transfer(amount Money, payer, payee Account) error {
	payer.Lock()
	cAmount := amount
	payer.Deposit(cAmount)
	err := payer.Withdraw(amount)
	if err != nil {
		return err
	}
	payer.Unlock()
	return nil
}

// Market errors

type invalidGoodTypeErr struct{}

func (e *invalidGoodTypeErr) Error() string {
	return "Good is the wrong type"
}

type noSellerOffersErr struct{}

func (e *noSellerOffersErr) Error() string {
	return "No seller offers available"
}

type marketDemandSatisfiedErr struct{}

func (e *marketDemandSatisfiedErr) Error() string {
	return "Market demand is satisified"
}
