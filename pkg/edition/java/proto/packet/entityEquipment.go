package packet

import (
	"bufio"
	"io"

	"github.com/Tnze/go-mc/nbt"
	"go.minekube.com/gate/pkg/edition/java/proto/util"
	"go.minekube.com/gate/pkg/gate/proto"
)

type EntityEquipment struct {
	EntityID int
	Slot     int16 // Equipment slot. 0: held, 1–4: armor slot (1: boots, 2: leggings, 3: chestplate, 4: helmet)
	Item     Slot
}

// https://minecraft.wiki/w/Java_Edition_protocol/Slot_data?oldid=2768577
type Slot struct {
	ID         int16
	ItemCount  byte
	ItemDamage int16
	NBT        map[string]interface{}
}

func (s *Slot) encode(wr io.Writer) error {
	if err := util.WriteInt16(wr, s.ID); err != nil {
		return err
	}
	if s.ID == -1 {
		return nil
	}
	if err := util.WriteByte(wr, s.ItemCount); err != nil {
		return err
	}
	if err := util.WriteInt16(wr, s.ItemDamage); err != nil {
		return err
	}
	if s.NBT == nil {
		if err := util.WriteByte(wr, 0); err != nil {
			return err
		}
		return nil
	}
	enc := nbt.NewEncoder(wr)
	enc.NetworkFormat(true)
	if err := enc.Encode(s.NBT, ""); err != nil {
		return err
	}
	return nil
}

func decodeSlot(rd io.Reader) (*Slot, error) {
	s := new(Slot)
	var err error

	s.ID, err = util.ReadInt16(rd)
	if err != nil {
		return nil, err
	}
	if s.ID == -1 {
		return s, nil
	}

	s.ItemCount, err = util.ReadByte(rd)
	if err != nil {
		return nil, err
	}
	s.ItemDamage, err = util.ReadInt16(rd)
	if err != nil {
		return nil, err
	}

	br := bufio.NewReader(rd)

	firstByte, err := br.Peek(1)
	if err != nil {
		return nil, err
	}
	if firstByte[0] == 0 {
		return s, nil
	}

	dec := nbt.NewDecoder(br)
	dec.NetworkFormat(true)
	_, err = dec.Decode(&s.NBT)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (j *EntityEquipment) Encode(c *proto.PacketContext, wr io.Writer) error {
	if err := util.WriteVarInt(wr, j.EntityID); err != nil {
		return err
	}
	if err := util.WriteInt16(wr, j.Slot); err != nil {
		return err
	}
	if err := util.WriteInt16(wr, j.Item.ID); err != nil {
		return err
	}
	return j.Item.encode(wr)
}

func (j *EntityEquipment) Decode(c *proto.PacketContext, rd io.Reader) (err error) {
	j.EntityID, err = util.ReadVarInt(rd)
	if err != nil {
		return err
	}
	j.Slot, err = util.ReadInt16(rd)
	if err != nil {
		return err
	}
	item, err := decodeSlot(rd)
	if err != nil {
		return err
	}
	j.Item = *item
	return nil
}
