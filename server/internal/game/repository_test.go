package game

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMemoryRepositoryReturnsIndependentSkillMaps(t *testing.T) {
	repository := NewMemoryPlayerRepository()
	player, err := NewPlayer("p-1", "Wayfarer", time.Now())
	if err != nil { t.Fatal(err) }
	if err := repository.Create(context.Background(), player); err != nil { t.Fatal(err) }

	loaded, err := repository.Get(context.Background(), player.ID)
	if err != nil { t.Fatal(err) }
	loaded.Skills["woodcutting"] = SkillProgress{Experience: 999, Level: 99}
	again, err := repository.Get(context.Background(), player.ID)
	if err != nil { t.Fatal(err) }
	if again.Skills["woodcutting"].Experience != 0 || again.Skills["woodcutting"].Level != 1 {
		t.Fatalf("mutating a returned player changed repository state: %+v", again.Skills["woodcutting"])
	}
}

func TestConcurrentExperienceAwardsAreNotLost(t *testing.T) {
	repository := NewMemoryPlayerRepository()
	player, err := NewPlayer("p-1", "Wayfarer", time.Now())
	if err != nil { t.Fatal(err) }
	if err := repository.Create(context.Background(), player); err != nil { t.Fatal(err) }

	const awards = 100
	var workers sync.WaitGroup
	workers.Add(awards)
	for i := 0; i < awards; i++ {
		go func() {
			defer workers.Done()
			if _, _, err := repository.AwardExperience(context.Background(), player.ID, "woodcutting", 1); err != nil {
				t.Errorf("AwardExperience() error = %v", err)
			}
		}()
	}
	workers.Wait()
	updated, err := repository.Get(context.Background(), player.ID)
	if err != nil { t.Fatal(err) }
	if got := updated.Skills["woodcutting"].Experience; got != awards {
		t.Fatalf("experience = %d, want %d", got, awards)
	}
}

func TestMemoryRepositoryReturnsIndependentBankAndInventoryMaps(t *testing.T) {
 repository:=NewMemoryPlayerRepository()
 player,err:=NewPlayer("p-bank","BankWayfarer",time.Now());if err!=nil{t.Fatal(err)}
 player.Inventory["softwood_log"]=3
 player.Bank["rough_stone"]=7
 if err:=repository.Create(context.Background(),player);err!=nil{t.Fatal(err)}
 loaded,err:=repository.Get(context.Background(),player.ID);if err!=nil{t.Fatal(err)}
 loaded.Bank["rough_stone"]=0
 loaded.Inventory["softwood_log"]=0
 again,err:=repository.Get(context.Background(),player.ID);if err!=nil{t.Fatal(err)}
 if again.Bank["rough_stone"]!=7||again.Inventory["softwood_log"]!=3{t.Fatalf("mutating returned bank/inventory changed repository state: bank=%v inventory=%v",again.Bank,again.Inventory)}
}
