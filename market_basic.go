package market

import (
	"errors"
	"sort"
)

type BasicMarket struct {
	goodType     GoodType       // type of good handled by the market
	maxOffer     Money          // maximum unit price the market is willing to pay per good
	demand       int            // number of goods the market wants to buy
	bought       []*Good        // goods bought by the market
	account      Account        // market's account
	sellerOffers []*SellerOffer // goods offered to market
	currentOffer *MarketCounter // market's current favoured offer (Could be a buffered channel?)
}

// NewBasicMarket constructor for BasicMarket.
func NewBasicMarket(goodType GoodType, maxOffer, demand int, balance Money) *BasicMarket {
	return &BasicMarket{
		goodType: GoodType(goodType),
		maxOffer: Money(maxOffer),
		demand:   demand,
		account:  NewSellerAccount(balance),
	}
}

// AddOffer adds a new SellerOffer to the market.
func (m *BasicMarket) AddOffer(o *SellerOffer) error {
	if len(o.goods) != 1 {
		return errors.New("Basic market only allows SellerOffers with a single good")
	}
	if o.goods[0].Type != m.goodType {
		return &invalidGoodTypeErr{}
	}

	m.sellerOffers = append(m.sellerOffers, o)
	return nil
}

// ConsiderOffers returns the latest market offer (or counter offer)
// for SellerOffers and moves the offer pointer forward.
func (m *BasicMarket) ConsiderOffers() (*MarketCounter, error) {

	var marketCounter Money

	if len(m.sellerOffers) == 0 {
		return nil, &noSellerOffersErr{}
	}

	if len(m.bought) >= m.demand {
		return nil, &marketDemandSatisfiedErr{}
	}

	// sort using Stable to ensure members with equal value
	// remain in place
	sort.Stable(ByOfferPrice(m.sellerOffers))

	// If the seller's price is greater than the max the market
	// is willing to pay, then counter the offer with the max
	// price, otherwise accept the seller's price
	if m.sellerOffers[0].price > m.maxOffer {
		marketCounter = m.maxOffer
	} else {
		marketCounter = m.sellerOffers[0].price
	}

	// If the market can't afford the seller's price or its own
	// max unit price then reset counter with balance
	if marketCounter > m.account.Balance() {
		marketCounter = m.account.Balance()
	}

	// Store the current offer for when a transaction is initiated
	m.currentOffer = &MarketCounter{
		price:       marketCounter,
		sellerOffer: m.sellerOffers[0],
	}

	// Shift the the current/best offer from the slice of available
	// seller offers
	m.sellerOffers = m.sellerOffers[1:]

	return m.currentOffer, nil
}

// TransactOffer handles the transaction between a seller (player) and
// a market.
func (m *BasicMarket) TransactOffer(c *MarketCounter) error {
	if m.currentOffer == nil {
		return errors.New("No offers available")
	}

	if m.currentOffer.sellerOffer != c.sellerOffer {
		return errors.New("Market offer is no longer current")
	}

	if len(m.bought) >= m.demand {
		return &marketDemandSatisfiedErr{}
	}

	err := Transfer(c.price, m.account, c.sellerOffer.account)
	if err != nil {
		return err
	}
	m.bought = append(m.bought, c.sellerOffer.goods...)
	// Remove the current offer
	m.currentOffer = nil

	return nil
}
