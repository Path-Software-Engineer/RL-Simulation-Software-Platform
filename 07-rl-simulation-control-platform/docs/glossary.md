# Reinforcement-learning product glossary

- **State:** the agent's observed row and column in the registered Gridworld.
- **Action:** one member of the allowlisted `up`, `right`, `down`, `left` action space.
- **Reward:** a numeric teaching signal returned by one transition; not a safety score.
- **Episode:** an ordered sequence from reset until goal, truncation or cancellation.
- **Policy:** a versioned mapping from observed state to action with hash and provenance.
- **Terminal:** the environment goal was reached. **Truncated** means the step boundary ended first.
- **Transition:** `(state, action, next state, reward, terminated, truncated)` at one step index.
- **Observed:** persisted evidence produced inside the controlled Gridworld simulation. It is not
  evidence from a physical environment and it is not a prediction.

Invariants: coordinates stay inside the 6×6 grid; obstacles cannot be entered; one episode has
strictly increasing step indexes; cumulative reward equals the sum of its persisted transitions;
only registered, compatible policy/environment pairs execute.
