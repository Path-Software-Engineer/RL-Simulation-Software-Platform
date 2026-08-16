FROM python:3.12.13-slim-bookworm
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 PIP_DISABLE_PIP_VERSION_CHECK=1
RUN useradd --create-home --uid 10001 runner
WORKDIR /workspace
COPY workers/rl-runner/requirements.lock /tmp/requirements.lock
RUN python -m pip install --no-cache-dir --requirement /tmp/requirements.lock
COPY workers/rl-runner/ ./workers/rl-runner/
COPY artifacts/ ./artifacts/
RUN python -m pip install --no-cache-dir --no-deps ./workers/rl-runner
USER runner
CMD ["python", "-m", "rl_runner.consumer"]
