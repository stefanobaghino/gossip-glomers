### Grow-Only Counter

Source: https://fly.io/dist-sys/4/

Reuses bits and pieces from the previous solutions: every add request that
comes into the system is marked with a unique ID and broadcasted to the rest
of the network (with some retries going on to recover from network
partitions). Each node keeps a map from a unique ID representing a specific
request to the delta associated with it and computes the sum based on those
events on the fly. Since the node is handled by a single goroutine, there's
no synchronization needed to access the node's state.
