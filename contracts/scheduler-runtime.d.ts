export interface SchedulerRequest {
  requestId: string;
  groupId: string;
  protocol: string;
  requestedModel: string;
  session?: {
    id?: string;
    conversationId?: string;
    promptCacheKey?: string;
    parentSessionId?: string;
    clientRequestId?: string;
    derivedHash?: string;
  };
}

export interface Candidate {
  instanceId: string;
  templateId: string;

  availability: number;
  decrease: number;

  activeRequests: number;
  effectiveAvailability: number;

  clientModel: string;
  upstreamModel: string;

  hardMaxConcurrency?: number;

  attributes: Record<string, unknown>;
}

export interface SchedulerContext {
  request: SchedulerRequest;
  groupConfig: Readonly<Record<string, unknown>>;
  groupState: Readonly<Record<string, unknown>>;

  patchGroupState(
    patch: Record<string, unknown>
  ): Promise<void>;

  clock: { now(): number };
}

export type SchedulerResult =
  | { selectedInstanceId: string }
  | { selectedInstanceId: null; reason: string };

export declare function select(
  ctx: SchedulerContext,
  candidates: ReadonlyArray<Candidate>
): Promise<SchedulerResult>;
