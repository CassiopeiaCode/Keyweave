export const manifest = {
  id: "request-rewrite-example",
  apiVersion: "ptpg.template/v1",
  protocols: ["anthropic.messages"],
  capabilities: { credentialBoundNetwork: true }
};

export async function execute(ctx) {
  const key = await ctx.secrets.get("key");
  if (!key) throw new Error("key missing");

  const body = await ctx.request.json();

  body.model = ctx.route.upstreamModel;
  body.metadata = {
    ...(body.metadata ?? {}),
    routed_by: "ptpg"
  };

  const headers = {
    "content-type": "application/json",
    "anthropic-version": "2023-06-01",
    "x-api-key": key.reveal()
  };

  const session = ctx.request.session?.id;
  if (session) headers["x-upstream-session"] = session;

  return ctx.http.request({
    url: String(ctx.fields.baseUrl).replace(/\/$/, "") + "/v1/messages",
    method: "POST",
    headers,
    body: JSON.stringify(body)
  });
}
