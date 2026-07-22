package seedsearch

import (
	"github.com/jonas/reaktor-sim/internal/board"
	"github.com/jonas/reaktor-sim/internal/finance"
	"github.com/jonas/reaktor-sim/internal/stats"
)

// CampaignMoneyFromChain estimates how each player's month budget was spent.
// Fresh budget is finance-card money per shift; carried savings add to later
// shift budgets but are not counted again in MonthBudget.
func CampaignMoneyFromChain(chain []Outcome, fin finance.Card) stats.CampaignMoney {
	if len(chain) == 0 {
		return stats.CampaignMoney{}
	}
	var cm stats.CampaignMoney
	cm.Shifts = len(chain)
	var prevBoard board.PlayerCosts
	for _, o := range chain {
		cm.MonthBudget[0] += fin.ReactorBudget
		cm.MonthBudget[1] += fin.GridBudget

		shiftBudgetP1 := fin.ReactorBudget + o.StartLeftover.Player1
		shiftBudgetP2 := fin.GridBudget + o.StartLeftover.Player2
		spentP1 := shiftBudgetP1 - o.EndLeftover.Player1 - o.RepairSpentP1
		spentP2 := shiftBudgetP2 - o.EndLeftover.Player2 - o.RepairSpentP2
		deltaP1 := o.BoardCosts.Player1 - prevBoard.Player1
		deltaP2 := o.BoardCosts.Player2 - prevBoard.Player2
		if rebuild := spentP1 - deltaP1; rebuild > 0 {
			cm.RebuildF[0] += float64(rebuild)
		}
		if rebuild := spentP2 - deltaP2; rebuild > 0 {
			cm.RebuildF[1] += float64(rebuild)
		}
		cm.AvgRepairF[0] += float64(o.RepairSpentP1)
		cm.AvgRepairF[1] += float64(o.RepairSpentP2)
		prevBoard = o.BoardCosts
	}
	return cm
}
