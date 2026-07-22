package stats

// CampaignMoney summarizes money flow across a multi-shift month.
type CampaignMoney struct {
	Shifts      int
	MonthBudget [2]int     // total fresh budget from the finance card over all shifts
	RebuildF    [2]float64 // field removal / overbuild overhead not visible in net board costs
	AvgRepairF  [2]float64 // total repair spend over all shifts [0]=P1, [1]=P2
}
