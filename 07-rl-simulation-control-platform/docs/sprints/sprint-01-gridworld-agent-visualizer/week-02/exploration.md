# Week 2 exploration

Questions resolved: a fixed 6×6 discrete environment keeps the product explainable; Q-Learning and
SARSA use JSON policy maps instead of notebook or pickle imports; rewards are `step=-0.04`,
`collision=-1`, `goal=10`; Redis carries versioned envelopes while Postgres remains authoritative.

Rejected alternatives: dynamic Python module names, uploaded pickles, hidden policy derivation and
running episodes inside the API request.
