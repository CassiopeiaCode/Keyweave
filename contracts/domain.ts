export type ProtocolId =
  | "openai.chat"
  | "openai.responses"
  | "anthropic.messages"
  | string;

export interface ModelDefinition {
  id: string;
  upstream?: string;
  displayName?: string;
  protocols?: ProtocolId[];
  forceResponseMapping?: boolean;
  inputModalities?: string[];
  outputModalities?: string[];
  metadata?: Record<string, unknown>;
}

export interface CredentialOfficialFields {
  address?: string | null;
  baseUrl?: string | null;
  models?: ModelDefinition[] | null;
  availability: number;
  proxy?: string | null;
}

export interface CredentialTemplate {
  id: string;
  name: string;
  description?: string;
  protocols: ProtocolId[];
  officialDefaults: Partial<CredentialOfficialFields>;
  customFields: Record<string, unknown>;
  state: Record<string, unknown>;
  jsSource: string;
  version: number;
}

export interface CredentialInstance extends CredentialOfficialFields {
  id: string;
  templateId: string;
  name?: string;
  keySecretRef?: string;
  customFields: Record<string, unknown>;
  state: Record<string, unknown>;
  version: number;
}

export interface CredentialGroup {
  id: string;
  name: string;
  templateIds: string[];
  schedulerType: string;
  schedulerCode?: string;
  config: Record<string, unknown>;
  state: Record<string, unknown>;
  version: number;
}

export interface AccessKey {
  id: string;
  name: string;
  secretHash: string;
  secretPrefix?: string;
  groupId: string;
  revokedAt?: string | null;
}

export interface CredentialSource {
  id: string;
  name: string;
  type: string;
  templateId?: string | null;
  config: Record<string, unknown>;
  code?: string;
  state: Record<string, unknown>;
  version: number;
}
