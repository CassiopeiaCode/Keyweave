// Conceptual library template. A production template loader may implement
// composition/mixins instead of ES module relative imports.
import * as generic from "./generic-openai.js";

export const manifest = {
  ...generic.manifest,
  id: "openrouter"
};

export function getAvailability(ctx) {
  const cooldownUntil = Number(ctx.fields.cooldownUntil ?? 0);
  if (cooldownUntil > ctx.clock.now()) return -1_000_000_000;
  return Number(ctx.instance.availability ?? 100);
}

export async function execute(ctx) {
  return generic.execute(ctx);
}

export async function afterResponse(ctx, response) {
  if (response.status === 429) {
    await ctx.instance.patch({
      customFields: {
        cooldownUntil: ctx.clock.now() + 30_000
      }
    });
  }
}

export async function onError(ctx, error) {
  const status = Number(error?.upstreamStatus ?? error?.status ?? 0);
  if (status === 429) {
    return { action: "retry", target: "another", cooldownMs: 30_000 };
  }
  return generic.onError(ctx, error);
}
