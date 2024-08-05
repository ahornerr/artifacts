package main

import (
	"context"
	"fmt"
	"github.com/promiseofcake/artifactsmmo-go-client/client"
	"time"
)

type LoopStopper func() bool

func NoopLoopStopper() bool {
	return false
}

func (r *Runner) GetNewTaskIfNecessary(ctx context.Context) error {
	if r.Char.Character.Task != "" {
		return nil
	}

	r.reportAction("Moving to taskmaster")
	err := r.Char.MoveClosest(ctx, r.Game.GetTaskMasterLocations("monsters"))
	if err != nil {
		return fmt.Errorf("moving to taskmaster: %w", err)
	}

	r.reportAction("Getting new task")
	_, err = r.Char.NewTask(ctx)
	if err != nil {
		return fmt.Errorf("getting new task: %w", err)
	}

	return nil
}

func (r *Runner) fightMonsterLoop(ctx context.Context, monsterName string, quantity int) error {
	numKilled := 0
	for numKilled < quantity {
		result, err := r.fightMonster(ctx, monsterName)
		if err != nil {
			return fmt.Errorf("killing monster: %w", err)
		}
		if result.Result == client.Win {
			numKilled++
		}
	}

	return nil
}

func (r *Runner) fightMonster(ctx context.Context, monsterName string) (*client.FightSchema, error) {
	monsterLocations := r.Game.GetMonsterLocations(monsterName)
	weaknesses := r.Game.GetMonster(monsterName).Resistance

	currentWeapon := r.Char.Character.WeaponSlot
	var currentWeaponBestStrength int
	if currentWeapon != "" {
		for element, resistance := range weaknesses {
			strength := r.Game.Items[currentWeapon].Attack[element] - resistance
			if strength > currentWeaponBestStrength {
				currentWeaponBestStrength = strength
			}
		}
	}

	bestBankItem, bestBankAttack := r.Game.BestUsableBankWeapon(r.Char.Character.Level, weaknesses)

	if bestBankAttack > currentWeaponBestStrength && currentWeapon != bestBankItem {
		r.reportAction("Moving to bank")
		err := r.Char.MoveClosest(ctx, r.Game.GetBankLocations())
		if err != nil {
			return nil, fmt.Errorf("moving to bank: %w", err)
		}

		// TODO: Also support armor

		if currentWeapon != "" {
			err := r.Char.Unequip(ctx, client.UnequipSchemaSlotWeapon)
			if err != nil {
				return nil, fmt.Errorf("unequipping slot: %w", err)
			}

			_, err = r.Char.DepositBank(ctx, currentWeapon, 1)
			if err != nil {
				return nil, fmt.Errorf("depositing weapon in bank: %w", err)
			}
		}

		_, err = r.Char.WithdrawBank(ctx, bestBankItem, 1)
		if err != nil {
			return nil, fmt.Errorf("withdrawing weapon from bank: %w", err)
		}

		err = r.Char.Equip(ctx, client.EquipSchemaSlotWeapon, bestBankItem)
		if err != nil {
			return nil, fmt.Errorf("equipping slot: %w", err)
		}
	}

	r.Char.WaitForCooldown()

	err := r.BankIfInventoryFull(ctx)
	if err != nil {
		return nil, err
	}

	if !r.Char.IsAtOneOf(monsterLocations) {
		r.reportAction("Moving to %s", monsterName)
		err := r.Char.MoveClosest(ctx, monsterLocations)
		if err != nil {
			return nil, fmt.Errorf("moving to %s: %w", monsterName, err)
		}
	}

	r.Char.WaitForCooldown()
	r.reportAction("Fighting %s", monsterName)

	fight, err := r.Char.Fight(ctx)
	if err != nil {
		return nil, fmt.Errorf("fighting: %w", err)
	}

	if fight.Result == client.Lose {
		r.reportAction("Fight lost!")
	} else if len(fight.Drops) > 0 && fight.Gold > 0 {
		r.reportAction("%s dropped %v and %d gold", monsterName, fight.Drops, fight.Gold)
	} else if len(fight.Drops) > 0 {
		r.reportAction("%s dropped %v", monsterName, fight.Drops)
	} else if fight.Gold > 0 {
		r.reportAction("%s dropped %d gold", monsterName, fight.Gold)
	}

	return fight, nil
}

func (r *Runner) taskLoop(ctx context.Context, shouldStop LoopStopper) error {
	for {
		switch r.Char.Character.TaskType {
		case "":
			err := r.GetNewTaskIfNecessary(ctx)
			if err != nil {
				return err
			}
		case "crafts":
		case "monsters":
			for r.Char.Character.TaskProgress < r.Char.Character.TaskTotal {
				if shouldStop() {
					return nil
				}
				_, err := r.fightMonster(ctx, r.Char.Character.Task)
				if err != nil {
					return fmt.Errorf("killing monster: %w", err)
				}
			}

			r.reportAction("Moving to taskmaster to complete task")
			err := r.Char.MoveClosest(ctx, r.Game.GetTaskMasterLocations("monsters"))
			if err != nil {
				return fmt.Errorf("moving to taskmaster: %w", err)
			}

			r.reportAction("Completing task")
			reward, err := r.Char.CompleteTask(ctx)
			if err != nil {
				return fmt.Errorf("completing task: %w", err)
			}

			r.reportAction("Reward %d %s!", reward.Quantity, reward.Code)
			r.Char.WaitForCooldown()
			time.Sleep(5 * time.Second)
		case "resources":
		}

	}
	return nil
}
