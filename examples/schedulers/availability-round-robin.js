export async function select(ctx, candidates) {
  const eligible = candidates.filter(c => {
    if (
      c.hardMaxConcurrency != null &&
      c.activeRequests >= c.hardMaxConcurrency
    ) {
      return false;
    }
    return Number.isFinite(c.effectiveAvailability);
  });

  if (eligible.length === 0) {
    return {
      selectedInstanceId: null,
      reason: "no eligible credentials"
    };
  }

  const max = Math.max(
    ...eligible.map(c => c.effectiveAvailability)
  );

  const top = eligible
    .filter(c => c.effectiveAvailability === max)
    .sort((a, b) => a.instanceId.localeCompare(b.instanceId));

  const cursors = ctx.groupState.rrCursor ?? {};
  const bucket =
    `${ctx.request.protocol}:${ctx.request.requestedModel}`;

  const pos = Number(cursors[bucket] ?? 0) % top.length;
  const selected = top[pos];

  await ctx.patchGroupState({
    rrCursor: {
      ...cursors,
      [bucket]: (pos + 1) % top.length
    }
  });

  return { selectedInstanceId: selected.instanceId };
}
