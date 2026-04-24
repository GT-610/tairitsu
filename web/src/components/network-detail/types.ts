import type { IpAssignmentPool, Route } from '../../services/api'

export interface NetworkMemberDevice {
  id: string;
  name: string;
  description: string;
  authorized: boolean;
  ipAssignments: string[];
  clientVersion: string;
  activeBridge: boolean;
  noAutoAssignIps: boolean;
}

export interface MemberFormState {
  name: string;
  authorized: boolean;
  activeBridge: boolean;
  noAutoAssignIps: boolean;
  ipAssignments: string[];
}

export interface BasicSettingsDraft {
  name: string;
  description: string;
}

export interface IPv4SettingsDraft {
  subnet: string;
  autoAssign: boolean;
  pools: IpAssignmentPool[];
  poolStartDraft: string;
  poolEndDraft: string;
}

export interface IPv6SettingsDraft {
  subnet: string;
  customAssign: boolean;
  rfc4193: boolean;
  plane6: boolean;
  pools: IpAssignmentPool[];
  poolStartDraft: string;
  poolEndDraft: string;
}

export interface DNSSettingsDraft {
  domain: string;
  servers: string[];
  serverDraft: string;
}

export interface MulticastSettingsDraft {
  multicastLimit: number;
  enableBroadcast: boolean;
}

export interface ManagedRoutesSettingsDraft {
  routes: Route[];
  routeDraft: Route;
}
