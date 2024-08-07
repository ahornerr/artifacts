package main

//func (r *Executor) fightMonsterLoop(ctx context.Context, monsterName string, quantity int) error {
//	numKilled := 0
//	for numKilled < quantity {
//		result, err := r.fightMonster(ctx, monsterName)
//		if err != nil {
//			return fmt.Errorf("killing monster: %w", err)
//		}
//		if result.Result == client.Win {
//			numKilled++
//		}
//	}
//
//	return nil
//}

//	if fight.Result == client.Lose {
//		r.reportAction("Fight lost!")
//	} else if len(fight.Drops) > 0 && fight.Gold > 0 {
//		r.reportAction("%s dropped %v and %d gold", monsterName, fight.Drops, fight.Gold)
//	} else if len(fight.Drops) > 0 {
//		r.reportAction("%s dropped %v", monsterName, fight.Drops)
//	} else if fight.Gold > 0 {
//		r.reportAction("%s dropped %d gold", monsterName, fight.Gold)
//	}
