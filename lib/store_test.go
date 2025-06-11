package lib

import "testing"

func TestStore_RegisterAndGet(t *testing.T) {
	d := NewViewDefinition("test-view", []string{"club, player, goals"}, UnitYear)
	m := NewViewMapper(d)
	i := NewViewInstance(m)

	store := NewStore()

	t.Run("registers a new view successfully", func(t *testing.T) {
		err := store.Register(i)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("fails to register duplicate view", func(t *testing.T) {
		err := store.Register(i)
		if err == nil {
			t.Fatal("expected error on duplicate register, got nil as err")
		}
	})

	t.Run("retrieves registered view", func(t *testing.T) {
		got, err := store.Get(i.Mapper.D.Name)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if got != i {
			got.Print("got: retrieves registered view")
			i.Print("i: retrieves registered view")
			t.Fatal("expected same view, got different")
		}
	})

	t.Run("fails to retrieve unknown view", func(t *testing.T) {
		_, err := store.Get("non-existent")
		if err == nil {
			t.Fatal("expected error on non-existent get, got nil as err")
		}
	})
}
