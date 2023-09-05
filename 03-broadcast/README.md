### Broadcast

Sources:
- https://fly.io/dist-sys/3a/
- https://fly.io/dist-sys/3b/
- https://fly.io/dist-sys/3c/
- https://fly.io/dist-sys/3d/
- https://fly.io/dist-sys/3e/

A node is represented by a single goroutine and instead of using callbacks directly, each
handler function proxies underlying channels used by the goroutine to signal its progress.
The main idea behind this is that the broadcast and read operation need to synchronize
access to the data received so far, so instead of using locking I preferred keeping the
code linear and not requiring to think about concurrent callbacks going on.

When a broadcast is received, if the message was not already present, its added to the node's
set of known messages and added to an outbox as well. The outbox is a buffered channel that
is periodically flushed. When the outbox is flushed, all of the messages contained therein are
sent to every neighbor of the node as a batch. Batching helps reduce the number of operations
performed to gossip the message around the cluster. Up to the last part of the challenge,
every previously unknown message was immediately sent individually to the node's neighbors.

The topology provided by Maelstrom is ignored and a statically-defined star-shaped one is used
instead. Even without batching, using this topology allowed to pass the second-to-last challenge.

Gossiping is the only process that is delegated to a separate set of goroutines, one for every
neighbor to which the data is sent. The process sends out the data and automatically retries if
no answer is heard within a specified timeout.

According to the tests I ran the solution can pass _both_ the performance challenges under optimal
network conditions (i.e. less than 20 messages per operation, less than 400 ms median latency,
less than 600 ms maximum latency), as well as recover from network partitions (albeit with
degraded performance). Tweaking the timeouts to have flushing happen right below the testing
latency and have the gossip timeout set to double the testing latency (to account for the
roundtrip) was quite important to pass these tests. This solution is fine to pass the challenge
but would require more work to be able to adapt to changing network conditions.
