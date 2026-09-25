package mcpe

import (
	"github.com/cespare/xxhash/v2"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ClientBlobCache is the server side of the client blob cache (ClientCacheStatus): chunk data is
// sent as xxHash64 hashes of "blobs" (a sub-chunk, or a chunk's biomes), and the client asks for
// the blobs it doesn't have with ClientCacheBlobStatus.
//
// It has no PocketMine-MP counterpart: PocketMine-MP 5.44 ignores ClientCacheStatus and only
// supports clients up to 1.26.30. A 1.26.51 Windows client, which enables the cache, disconnects
// with "Block" when this server sends its request-mode chunks without the cache, while it plays
// fine on Dragonfly, which uses it; this follows Dragonfly's session/chunk.go and
// handler_client_cache_blob_status.go.
type ClientBlobCache struct {
	blobs map[uint64][]byte
}

// maxPendingBlobs is how many blobs the client may leave unacknowledged (Dragonfly's limit).
const maxPendingBlobs = 4096

func NewClientBlobCache() *ClientBlobCache {
	return &ClientBlobCache{blobs: map[uint64][]byte{}}
}

// Track stores blob until the client reports it as a hit or a miss, and returns its hash. ok is
// false when too many blobs are pending, in which case the data should be sent without the cache.
func (c *ClientBlobCache) Track(blob []byte) (hash uint64, ok bool) {
	if len(c.blobs) > maxPendingBlobs {
		return 0, false
	}
	hash = xxhash.Sum64(blob)
	c.blobs[hash] = blob
	return hash, true
}

// HandleBlobStatus answers a ClientCacheBlobStatus: blobs the client already has are forgotten,
// and the ones it's missing are sent. It returns nil when there's nothing to send.
func (c *ClientBlobCache) HandleBlobStatus(pk *packet.ClientCacheBlobStatus) *packet.ClientCacheMissResponse {
	for _, hit := range pk.HitHashes {
		delete(c.blobs, hit)
	}
	resp := &packet.ClientCacheMissResponse{Blobs: make([]protocol.CacheBlob, 0, len(pk.MissHashes))}
	for _, miss := range pk.MissHashes {
		blob, ok := c.blobs[miss]
		if !ok {
			// Expected when the same blob was sent several times in a short time.
			continue
		}
		resp.Blobs = append(resp.Blobs, protocol.CacheBlob{Hash: miss, Payload: blob})
		delete(c.blobs, miss)
	}
	if len(resp.Blobs) == 0 {
		return nil
	}
	return resp
}
