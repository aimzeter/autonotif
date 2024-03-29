package autonotif

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aimzeter/autonotif/config"
	"github.com/aimzeter/autonotif/entity"
)

type Autonotif struct {
	d *Dependencies
}

func (a *Autonotif) Terminate() error {
	return nil
}

func (a *Autonotif) HealthCheck() error {
	return nil
}

func (a *Autonotif) Run() error {
	for chainType, chainConf := range a.d.conf.ChainList {
		log.Printf("INFO | chain %s running...\n", strings.ToUpper(chainType))

		err := a.notifyRecentProposal(chainType, chainConf)
		if err != nil {
			log.Printf("ERROR | chain %s runner.notifyRecentProposal: %s\n", strings.ToUpper(chainType), err)
			continue
		}

		log.Printf("INFO | chain %s run successfully\n", strings.ToUpper(chainType))
	}

	return nil
}

func (a *Autonotif) notifyRecentProposal(chainType string, chainConf config.Chain) error {
	ctx := context.Background()

	lastID, err := a.d.dsStore.GetLastID(ctx, chainType)
	if err != nil {
		return fmt.Errorf("dsStore.GetLastID: %s", err.Error())
	}

	nextID := lastID + 1
	p := &entity.Proposal{
		ID:          nextID,
		ChainType:   chainType,
		ChainConfig: chainConf,
	}

	p, err = a.d.dsAPI.GetProposalDetail(ctx, p)
	if err == entity.ErrProposalNotYetExistInDatasource {
		return nil
	}

	if err != nil {
		return fmt.Errorf("dsAPI.GetProposalDetail: %s", err.Error())
	}

	err = a.notify(ctx, p)
	if err != nil {
		return fmt.Errorf("notify: %s", err.Error())
	}

	err = a.d.dsStore.Set(ctx, p)
	if err != nil {
		return fmt.Errorf("dsStore.Set: %s", err.Error())
	}

	return nil
}

func (a *Autonotif) notify(ctx context.Context, p *entity.Proposal) error {
	if !p.IsShouldNotify() {
		return nil
	}

	err := a.d.notifier.SendMessage(ctx, p)
	if err != nil {
		return fmt.Errorf("notifier.SendMessage: %s", err.Error())
	}

	return nil
}
