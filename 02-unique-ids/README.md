### Unique IDs

Source: https://fly.io/dist-sys/2/

By prepending the node identifier to a sequential ID, the node is able to generate globally unique identifiers in the face of network partitions.

The assumptions that make this approach work are that the nodes do not crash (which would require persisting the latest emitted sequential ID) and that the topology (i.e. the node names) is known and stable (if two nodes "swap" names while running, this approach fails).