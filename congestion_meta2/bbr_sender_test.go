package congestion

import (
	"testing"
	"time"

	"github.com/sagernet/quic-go/congestion"
	"github.com/sagernet/quic-go/monotime"
)

func TestAckOnlyTrafficDoesNotStarveDataPacing(t *testing.T) {
	t.Parallel()

	for _, profile := range []Profile{ProfileConservative, ProfileStandard, ProfileAggressive} {
		t.Run(profile.name, func(t *testing.T) {
			t.Parallel()

			b := NewBbrSenderWithProfile(congestion.InitialPacketSize, profile)
			b.pacingRate = Bandwidth(minBps) * BytesPerSecond
			now := monotime.Now()
			if !b.HasPacingBudget(now) {
				t.Fatal("new pacer has no data budget")
			}
			for packet := congestion.PacketNumber(0); packet < 400; packet++ {
				now = now.Add(100 * time.Microsecond)
				b.OnPacketSent(now, 0, packet, 100, false)
			}
			if !b.HasPacingBudget(now) {
				t.Fatal("ACK-only traffic must leave data pacing available")
			}
			budget := b.pacer.Budget(now)
			b.OnPacketSent(now, 0, 400, b.maxDatagramSize, true)
			if got, want := b.pacer.Budget(now), budget-b.maxDatagramSize; got != want {
				t.Fatalf("data pacing budget = %d; want %d", got, want)
			}
		})
	}
}
