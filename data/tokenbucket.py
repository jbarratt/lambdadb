import time
from threading import Lock


class TokenBucket(object):

    def __init__(self, rate=1, tokens=0, capacity=100):
        # immutable attributes
        self.lock = Lock()
        self.rate = rate
        self.capacity = capacity
        # mutable attributes
        self._tokens = tokens
        self._time = time.monotonic()

    def _adjust(self):
        """
        Update internal time and tokens
        """
        now = time.monotonic()
        elapsed = now - self._time

        self._tokens = min(
            self.capacity,
            self._tokens + elapsed*self.rate
        )
        self._time = now

    def tokens(self):
        """
        Publicly accessible view of how many tokens the bucket has.
        """
        with self.lock:
            self._adjust()
            return self._tokens

    def consume(self, tokens):
        """
        Consume `tokens` tokens from the bucket, blocking until they are
        available
        """
        with self.lock:
            self._adjust()
            self._tokens -= tokens
            if self._tokens > 0:
                return
            else:
                to_sleep = -self._tokens/self.rate
                time.sleep(to_sleep)
                self._adjust()
                assert self._tokens >= 0
