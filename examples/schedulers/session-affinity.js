function sessionKey(request) {
  const s = request.session ?? {};
  return (
    s.id ??
    s.conversationId ??
    s.promptCacheKey ??
    s.parentSessionId ??
    s.clientRequestId ??
    s.derivedHash ??
    null
  );
}

export async function select(ctx, candidates) {
  const now = ctx.clock.now();
  const key = sessionKey(ctx.request);
  const ttlMs =
    Number(ctx.groupConfig.sessionAffinityTtlMs ?? 3_600_000);

  const bindings = ctx.groupState.bindings ?? {};

  if (key) {
    const binding = bindings[key];
    if (binding && binding.expiresAt > now) {
      const existing = candidates.find(
        c =>
          c.instanceId === binding.instanceId &&
          (c.hardMaxConcurrency == null ||
           c.activeRequests < c.hardMaxConcurrency)
      );
      if (existing) {
        return { selectedInstanceId: existing.instanceId };
      }
    }
  }

  const eligible = candidates.filter(
    c =>
      c.hardMaxConcurrency == null ||
      c.activeRequests < c.hardMaxConcurrency
  );

  if (!eligible.length) {
    return {
      selectedInstanceId: null,
      reason: "no eligible credentials"
    };
  }

  eligible.sort(
    (a, b) =>
      b.effectiveAvailability - a.effectiveAvailability ||
      a.instanceId.localeCompare(b.instanceId)
  );

  const selected = eligible[0];

  if (key) {
    await ctx.patchGroupState({
      bindings: {
        ...bindings,
        [key]: {
          instanceId: selected.instanceId,
          expiresAt: now + ttlMs
        }
      }
    });
  }

  return { selectedInstanceId: selected.instanceId };
}
