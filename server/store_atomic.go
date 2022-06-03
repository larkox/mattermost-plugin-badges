package main

import (
	"encoding/json"
	"errors"

	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
)

const ATOMICRETRIES = 3

func (s *store) doAtomic(f func() (bool, error)) error {
	done := false
	for i := 0; i < ATOMICRETRIES; i++ {
		var err error
		done, err = f()
		if err != nil {
			return err
		}
		if done {
			break
		}
	}
	if !done {
		return errors.New("too many attempts on atomic retry")
	}

	return nil
}

func (s *store) compareAndSet(key string, old []byte, value interface{}) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	done, appErr := s.api.KVCompareAndSet(key, old, data)
	if appErr != nil {
		return false, appErr
	}

	return done, nil
}

func (s *store) atomicRemoveBadge(bID badgesmodel.BadgeID) (bool, error) {
	bb, data, err := s.getAllBadges()
	if err != nil {
		return false, err
	}

	for i, b := range bb {
		if b.ID == bID {
			bb = append(bb[:i], bb[i+1:]...)
			break
		}
	}

	return s.compareAndSet(KVKeyBadges, data, bb)
}

func (s *store) atomicRemoveBadgeFromOwnership(bID badgesmodel.BadgeID) (bool, error) {
	ownership, data, err := s.getOwnershipList()
	if err != nil {
		return false, err
	}

	toDelete := []int{}
	for i, o := range ownership {
		if o.Badge == bID {
			toDelete = append([]int{i}, toDelete...)
		}
	}

	for _, index := range toDelete {
		ownership = append(ownership[:index], ownership[index+1:]...)
	}

	return s.compareAndSet(KVKeyOwnership, data, ownership)
}

func (s *store) atomicAddBadge(b *badgesmodel.Badge) (bool, error) {
	bb, data, err := s.getAllBadges()
	if err != nil {
		return false, err
	}

	bb = append(bb, b)

	return s.compareAndSet(KVKeyBadges, data, bb)
}

func (s *store) atomicAddType(t *badgesmodel.BadgeTypeDefinition) (bool, error) {
	tt, data, err := s.getAllTypes()
	if err != nil {
		return false, err
	}

	tt = append(tt, t)

	return s.compareAndSet(KVKeyTypes, data, tt)
}

func (s *store) atomicAddBadgeToOwnership(o badgesmodel.Ownership, isMultiple bool) (shouldNotify bool, done bool, err error) {
	ownership, data, err := s.getOwnershipList()
	if err != nil {
		return false, false, err
	}

	if !isMultiple && ownership.IsOwned(o.User, o.Badge) {
		return false, true, nil
	}

	ownership = append(ownership, o)

	done, err = s.compareAndSet(KVKeyOwnership, data, ownership)
	return done, done, err
}

func (s *store) atomicUpdateType(t *badgesmodel.BadgeTypeDefinition) (bool, error) {
	tt, data, err := s.getAllTypes()
	if err != nil {
		return false, err
	}

	found := false
	for i, tOld := range tt {
		if tOld.ID == t.ID {
			tt[i] = t
			found = true
			break
		}
	}

	if !found {
		return false, errors.New("not found")
	}

	return s.compareAndSet(KVKeyTypes, data, tt)
}

func (s *store) atomicUpdateBadge(b *badgesmodel.Badge) (bool, error) {
	bb, data, err := s.getAllBadges()
	if err != nil {
		return false, err
	}

	found := false
	for i, bOld := range bb {
		if bOld.ID == b.ID {
			bb[i] = b
			found = true
			break
		}
	}
	if !found {
		return false, errors.New("not found")
	}

	return s.compareAndSet(KVKeyBadges, data, bb)
}

func (s *store) atomicAddSubscription(toAdd badgesmodel.Subscription) (bool, error) {
	subs, data, err := s.getAllSubscriptions()
	if err != nil {
		return false, err
	}

	for _, sub := range subs {
		if sub.ChannelID == toAdd.ChannelID && sub.TypeID == toAdd.TypeID {
			return true, nil
		}
	}

	subs = append(subs, toAdd)

	return s.compareAndSet(KVKeySubscriptions, data, subs)
}

func (s *store) atomicRemoveSubscription(toRemove badgesmodel.Subscription) (bool, error) {
	subs, data, err := s.getAllSubscriptions()
	if err != nil {
		return false, err
	}

	found := false
	for i, sub := range subs {
		if sub.ChannelID == toRemove.ChannelID && sub.TypeID == toRemove.TypeID {
			subs = append(subs[:i], subs[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return true, nil
	}

	return s.compareAndSet(KVKeySubscriptions, data, subs)
}
