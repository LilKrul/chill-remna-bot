package app

type uiState struct {
	buyMonths    int
	topUpKopecks int64
	awaitTopUp   bool
	awaitPromo   bool
	awaitShotReq int64
	rejectReq    int64

	adminInput  string
	priceMonths int

	welcomeAwait       string
	awaitSectionBanner string
	awaitEmojiFor      string

	inputBack string

	broadcastText string

	p2pSubmitMsgID int
	p2pShotMsgID   int
}

func (a *App) getUI(chatID int64) *uiState {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.ui == nil {
		a.ui = map[int64]*uiState{}
	}
	st := a.ui[chatID]
	if st == nil {
		st = &uiState{}
		a.ui[chatID] = st
	}
	return st
}
