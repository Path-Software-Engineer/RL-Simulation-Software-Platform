# Week 3 exploration

The player must distinguish visual playback from runner pause. Playback is local inspection of
persisted transitions; runner pause is a cooperative backend command. WebSocket is useful only as a
change notification because reconnects can lose ephemeral messages. REST therefore performs every
authoritative resync.
