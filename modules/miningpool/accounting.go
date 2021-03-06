package pool

import (
	"fmt"

	"github.com/NebulousLabs/Sia/modules"
)

func (p *Pool) BlockCount() uint64 {
	p.mu.RLock()
	defer func() {
		p.mu.RUnlock()
	}()
	return p.blockCounter
}

// processBlockPayout TODO: would move this part all to yiimp
func (p *Pool) processBlockPayout(block uint64) {
	// info := p.BlockInfo(block)
	// fmt.Printf("Processing block %d, split %d ways\n", block, len(info))
	// for _, bi := range info {
	// 	p.processClientPayout(block, bi)
	// }
	// p.markBlockPaid(block)

}

// processClientPayout modifies the client internal balance and records
// the transaction in the ledger
func (p *Pool) processClientPayout(block uint64, bi modules.PoolBlock) {
	client := p.FindClientDB(bi.ClientName)
	err := p.modifyClientBalance(client.cr.clientID, bi.ClientReward)
	if err != nil {
		p.log.Printf("Failed to modify client %s balance: %s\n", bi.ClientName, err)
		return
	}
	memo := fmt.Sprintf("Reward from block %d, percentage %2.2f", block, bi.ClientPercentage)
	err = p.makeClientTransaction(client.cr.clientID, bi.ClientReward, memo)
	if err != nil {
		p.log.Printf("Failed to create ledger entry for client %s: %s\n", bi.ClientName, err)
	}
}

func (p *Pool) processOperatorPayout() {

}
