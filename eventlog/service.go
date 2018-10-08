package eventlog

import (
	"errors"
	"fmt"
	"time"

	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
	"github.com/dedis/protobuf"
)

// ServiceName is the service name for the EventLog service.
var ServiceName = "EventLog"

var sid onet.ServiceID

func init() {
	var err error
	sid, err = onet.RegisterNewService(ServiceName, newService)
	if err != nil {
		log.Fatal(err)
	}
}

// Service is the EventLog service.
type Service struct {
	*onet.ServiceProcessor
	omni         *byzcoin.Service
	bucketMaxAge time.Duration
}

const defaultBlockInterval = 5 * time.Second

// This should be a const, but we want to be able to hack it from tests.
var searchMax = 10000

// Search will search the event log for matching entries.
func (s *Service) Search(req *SearchRequest) (*SearchResponse, error) {
	if req.ID.IsNull() {
		return nil, errors.New("skipchain ID required")
	}

	if req.To == 0 {
		req.To = time.Now().UnixNano()
	}

	v := s.omni.GetCollectionView(req.ID)
	el := &eventLog{Instance: req.Instance, v: v}

	id, b, err := el.getLatestBucket()
	if err != nil {
		return nil, err
	}
	if b == nil {
		// There are no events yet on this chain, so return no results.
		return &SearchResponse{}, nil
	}

	// bEnd is normally updated from the last bucket's start. For the latest
	// bucket, bEnd is now.
	bEnd := time.Now().UnixNano()

	// Walk backwards in the bucket chain through 2 zones: first where the
	// bucket covers time that is not in our search range, and then where the buckets
	// do cover the search range. When we see a bucket that ends before our search
	// range, we can stop walking buckets.
	var buckets []*bucket
	var bids [][]byte
	for {
		if req.From > bEnd {
			// This bucket is before the search range, so we are done walking back the bucket chain.
			break
		}

		if req.To < b.Start {
			// This bucket is after the search range, so we do not add it to buckets, but
			// we keep walking up the chain.
		} else {
			buckets = append(buckets, b)
			bids = append(bids, id)
		}

		if b.isFirst() {
			break
		}
		bEnd = b.Start
		id = b.Prev
		b, err = el.getBucketByID(id)
		if err != nil {
			// This indicates that the event log data structure is wrong, so
			// we cannot claim to correctly search it. Give up instead.
			log.Errorf("expected event log bucket id %v not found: %v", string(id), err)
			return nil, err
		}
	}

	reply := &SearchResponse{}

	// Process the time buckets from earliest to latest so that
	// if we truncate, it is the latest events that are not returned,
	// so that they can set req.From = resp.Events[len(resp.Events)-1].When.
filter:
	for i := len(buckets) - 1; i >= 0; i-- {
		b := buckets[i]
		for _, e := range b.EventRefs {
			ev, err := getEventByID(v, e)
			if err != nil {
				log.Errorf("bucket %x points to event %x, but the event was not found: %v", bids[i], e, err)
				return nil, err
			}

			if req.From <= ev.When && ev.When < req.To {
				if req.Topic == "" || req.Topic == ev.Topic {
					reply.Events = append(reply.Events, *ev)
					if len(reply.Events) >= searchMax {
						reply.Truncated = true
						break filter
					}
				}
			}
		}
	}

	return reply, nil
}

const contractName = "eventlog"

func (s *Service) decodeAndCheckEvent(coll byzcoin.CollectionView, eventBuf []byte) (*Event, error) {
	// Check the timestamp of the event: it should never be in the future,
	// and it should not be more than 30 seconds in the past. (Why 30 sec
	// and not something more auto-scaling like blockInterval * 30?
	// Because a # of blocks limit is too fragile when using fast blocks for
	// tests.)
	//
	// Also: An event a few seconds into the future is OK because there might be
	// time skew between a legitimate event producer and the network. See issue #1331.
	event := &Event{}
	err := protobuf.Decode(eventBuf, event)
	if err != nil {
		return nil, err
	}
	when := time.Unix(0, event.When)
	now := time.Now()
	if when.Before(now.Add(-30 * time.Second)) {
		return nil, fmt.Errorf("event timestamp too long ago - when=%v, now=%v", when, now)
	}
	if when.After(now.Add(5 * time.Second)) {
		return nil, errors.New("event timestamp is too far in the future")
	}
	return event, nil
}

// invoke will add an event and update the corresponding indices.
func (s *Service) invoke(v byzcoin.CollectionView, inst byzcoin.Instruction, c []byzcoin.Coin) (sc []byzcoin.StateChange, cOut []byzcoin.Coin, err error) {
	cOut = c

	_, cid, darcID, err := v.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return nil, nil, err
	}
	if cid != contractName {
		return nil, nil, fmt.Errorf("expected contract ID to be %s but got %s", contractName, cid)
	}

	eventBuf := inst.Invoke.Args.Search("event")
	if eventBuf == nil {
		return nil, nil, errors.New("expected a named argument of \"event\"")
	}

	event, err := s.decodeAndCheckEvent(v, eventBuf)
	if err != nil {
		return nil, nil, err
	}

	// Even though this is an invoke, we'll use the Spawn convention,
	// since the new event is essentially being spawned on this eventlog.
	eventID := inst.DeriveID("")

	sc = append(sc, byzcoin.NewStateChange(byzcoin.Create, eventID, cid, eventBuf, darcID))

	// Walk from latest bucket back towards beginning looking for the right bucket.
	//
	// If you don't find a bucket with b.Start <= ev.When,
	// create a new bucket, put in the event, set the start, emit the bucket,
	// update prev in the bucket before (and also possibly the index key).
	//
	// If you find an existing latest bucket, and b.Start is more than X seconds
	// ago, make a new bucket anyway.
	//
	// If you find the right bucket, add the event and emit the updated bucket.
	// For now: buckets are allowed to grow as big as needed (but the previous
	// rule prevents buckets from getting too big by timing them out).

	el := &eventLog{Instance: inst.InstanceID, v: v}
	bID, b, err := el.getLatestBucket()
	if err != nil {
		return nil, nil, err
	}
	isHead := true

	for b != nil && !b.isFirst() {
		if b.Start <= event.When {
			break
		}
		bID = b.Prev
		b, err = el.getBucketByID(bID)
		if err != nil {
			return nil, nil, err
		}
		isHead = false
	}

	// Make a new head bucket if:
	//   No latest bucket (b == nil).
	//     or
	//   Found a bucket, and it is head, and it is too old.
	if b == nil || isHead && time.Duration(event.When-b.Start) > s.bucketMaxAge {
		newBid := inst.DeriveID("bucket")

		if b == nil {
			// Special case: The first bucket for an eventlog
			// needs a catch-all bucket before it, in case later
			// events come in.
			catchID := inst.DeriveID("bucket-catch-all")
			newb := &bucket{
				Start:     0,
				Prev:      nil,
				EventRefs: nil,
			}
			buf, err := protobuf.Encode(newb)
			if err != nil {
				return nil, nil, err
			}
			sc = append(sc, byzcoin.NewStateChange(byzcoin.Create, catchID, cid, buf, darcID))
			bID = catchID.Slice()
		}

		newb := &bucket{
			// This new bucket will start with this event.
			Start: event.When,
			// It links to the previous latest bucket, or to the catch-all bucket
			// if there was no previous bucket.
			Prev:      bID,
			EventRefs: [][]byte{eventID.Slice()},
		}
		buf, err := protobuf.Encode(newb)
		if err != nil {
			return nil, nil, err
		}
		sc = append(sc, byzcoin.NewStateChange(byzcoin.Create, newBid, cid, buf, darcID))

		// Update the pointer to the latest bucket.
		sc = append(sc, byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID, cid, newBid.Slice(), darcID))
	} else {
		// Otherwise just add into whatever bucket we found, no matter how
		// many are already there. (Splitting buckets is hard and not important to us.)
		b.EventRefs = append(b.EventRefs, eventID.Slice())
		bucketBuf, err := protobuf.Encode(b)
		if err != nil {
			return nil, nil, err
		}
		sc = append(sc,
			byzcoin.StateChange{
				StateAction: byzcoin.Update,
				InstanceID:  bID,
				ContractID:  []byte(contractName),
				Value:       bucketBuf,
			})
	}
	return
}

func (s *Service) spawn(v byzcoin.CollectionView, inst byzcoin.Instruction, c []byzcoin.Coin) ([]byzcoin.StateChange, []byzcoin.Coin, error) {
	_, _, darcID, err := v.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return nil, nil, err
	}

	cid := inst.Spawn.ContractID
	if cid != contractName {
		return nil, nil, errors.New("invalid contract ID: " + cid)
	}

	// Store zeros as the pointer to the first bucket because there are not yet
	// any events in this event log.
	return []byzcoin.StateChange{
		byzcoin.NewStateChange(byzcoin.Create, inst.DeriveID(""), cid, make([]byte, 32), darcID),
	}, nil, nil
}

// contractFunction is the function that runs to process a transaction of
// type "eventlog"
func (s *Service) contractFunction(v byzcoin.CollectionView, inst byzcoin.Instruction, c []byzcoin.Coin) ([]byzcoin.StateChange, []byzcoin.Coin, error) {

	err := inst.VerifyDarcSignature(v)
	if err != nil {
		return nil, nil, err
	}

	switch inst.GetType() {
	case byzcoin.InvokeType:
		return s.invoke(v, inst, c)
	case byzcoin.SpawnType:
		return s.spawn(v, inst, c)
	default:
		return nil, nil, errors.New("invalid type")
	}
}

// newService receives the context that holds information about the node it's
// running on. Saving and loading can be done using the context. The data will
// be stored in memory for tests and simulations, and on disk for real
// deployments.
func newService(c *onet.Context) (onet.Service, error) {
	s := &Service{
		ServiceProcessor: onet.NewServiceProcessor(c),
		omni:             c.Service(byzcoin.ServiceName).(*byzcoin.Service),
		// Set a relatively low time for bucketMaxAge: during peak message arrival
		// this will pretect the buckets from getting too big. During low message
		// arrival (< 1 per 5 sec) it does not create extra buckets, because time
		// periods with no events do not need buckets created for them.
		bucketMaxAge: 5 * time.Second,
	}
	if err := s.RegisterHandlers(s.Search); err != nil {
		log.ErrFatal(err, "Couldn't register messages")
	}

	byzcoin.RegisterContract(s, contractName, s.contractFunction)
	return s, nil
}

func getEventByID(view byzcoin.CollectionView, eid []byte) (*Event, error) {
	r, err := view.Get(eid).Record()
	if err != nil {
		return nil, err
	}
	v, err := r.Values()
	if err != nil {
		return nil, err
	}
	newval, ok := v[0].([]byte)
	if !ok {
		return nil, errors.New("invalid value")
	}
	var e Event
	if err := protobuf.Decode(newval, &e); err != nil {
		return nil, err
	}
	return &e, nil
}
