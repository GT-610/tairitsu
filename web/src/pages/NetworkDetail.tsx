import { useState, useEffect } from 'react';
import { Box, Typography, Card, CardContent, CircularProgress, Alert, Button, Grid, IconButton, Tabs, Tab, Paper, TextField, Switch, FormControlLabel, Dialog, DialogTitle, DialogContent, DialogActions, DialogContentText, Snackbar, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Chip, Menu, MenuItem, Stack } from '@mui/material';
import { ArrowBack, ContentCopy, MoreHoriz } from '@mui/icons-material';
import { useParams, Link } from 'react-router-dom';
import { memberAPI, networkAPI, Network, NetworkConfig, NetworkMetadataUpdateRequest, IpAssignmentPool, Route, type Member as ApiMember } from '../services/api';
import { getErrorMessage } from '../services/errors';

// Default configuration for optional fields (null instead of default values)
const defaultNetworkConfig: Partial<NetworkConfig> = {
  private: true,
  allowPassiveBridging: true,
  enableBroadcast: true,
  mtu: 2800,
  multicastLimit: 32,
  v4AssignMode: {
    zt: false
  },
  v6AssignMode: {
    zt: false,
    '6plane': false,
    rfc4193: false
  }
};

interface NetworkMemberDevice {
  id: string;
  name: string;
  description: string;
  authorized: boolean;
  ipAssignments: string[];
  clientVersion: string;
  lastSeenLabel: string;
  activeBridge: boolean;
  noAutoAssignIps: boolean;
}

type ParsedIPv6CIDR = {
  network: bigint;
  prefix: number;
};

function parseIPv4Address(value: string): number | null {
  const parts = value.trim().split('.');
  if (parts.length !== 4) return null;

  let result = 0;
  for (const part of parts) {
    if (!/^\d+$/.test(part)) return null;
    const octet = Number(part);
    if (octet < 0 || octet > 255) return null;
    result = (result << 8) + octet;
  }

  return result >>> 0;
}

function parseIPv4CIDR(target: string): { network: number; mask: number } | null {
  const [address, prefixText] = target.trim().split('/');
  if (!address || !prefixText || !/^\d+$/.test(prefixText)) return null;

  const prefix = Number(prefixText);
  if (prefix < 0 || prefix > 32) return null;

  const ip = parseIPv4Address(address);
  if (ip === null) return null;

  const mask = prefix === 0 ? 0 : ((0xffffffff << (32 - prefix)) >>> 0);
  return {
    network: ip & mask,
    mask,
  };
}

function formatIPv4Address(value: number): string {
  return [
    (value >>> 24) & 255,
    (value >>> 16) & 255,
    (value >>> 8) & 255,
    value & 255,
  ].join('.');
}

function parseIPv6Address(value: string): bigint | null {
  const input = value.trim().toLowerCase();
  if (input === '' || !/^[0-9a-f:]+$/.test(input)) return null;
  if ((input.match(/::/g) || []).length > 1) return null;

  const [leftPart, rightPart] = input.split('::');
  const left: string[] = leftPart ? leftPart.split(':').filter((part) => part !== '') : [];
  const right: string[] = rightPart ? rightPart.split(':').filter((part) => part !== '') : [];

  if (!input.includes('::')) {
    if (left.length !== 8) return null;
  } else if (left.length + right.length >= 8) {
    return null;
  }

  const missing = input.includes('::') ? 8 - (left.length + right.length) : 0;
  const filledGroups: string[] = Array.from({ length: missing }, () => '0');
  const groups: string[] = input.includes('::') ? [...left, ...filledGroups, ...right] : left;
  if (groups.length !== 8) return null;

  let result = 0n;
  for (const group of groups) {
    if (!/^[0-9a-f]{1,4}$/.test(group)) return null;
    result = (result << 16n) + BigInt(parseInt(group, 16));
  }

  return result;
}

function formatIPv6Address(value: bigint): string {
  const groups: string[] = [];
  for (let i = 7; i >= 0; i -= 1) {
    const group = Number((value >> BigInt(i * 16)) & 0xffffn);
    groups.push(group.toString(16));
  }
  return groups.join(':');
}

function parseIPv6CIDR(target: string): ParsedIPv6CIDR | null {
  const [address, prefixText] = target.trim().split('/');
  if (!address || !prefixText || !/^\d+$/.test(prefixText)) return null;

  const prefix = Number(prefixText);
  if (prefix < 0 || prefix > 128) return null;

  const ip = parseIPv6Address(address);
  if (ip === null) return null;

  const mask = prefix === 0 ? 0n : ((1n << BigInt(prefix)) - 1n) << BigInt(128 - prefix);
  return {
    network: ip & mask,
    prefix,
  };
}

function normalizeIPv6CIDR(target: string): string | null {
  const parsed = parseIPv6CIDR(target);
  if (!parsed) return null;
  return `${formatIPv6Address(parsed.network)}/${parsed.prefix}`;
}

function normalizeIPv4CIDR(target: string): string | null {
  const parsed = parseIPv4CIDR(target);
  if (!parsed) return null;

  const prefix = Number(target.trim().split('/')[1]);
  return `${formatIPv4Address(parsed.network)}/${prefix}`;
}

function getCIDRPrefix(target: string): number | null {
  const prefixText = target.trim().split('/')[1];
  if (!prefixText || !/^\d+$/.test(prefixText)) return null;
  const prefix = Number(prefixText);
  if (prefix < 0 || prefix > 32) return null;
  return prefix;
}

function getPrimaryIPv4Subnet(routes: Route[]): string {
  const primaryRoute = routes.find((route) => !route.via && parseIPv4CIDR(route.target));
  return primaryRoute?.target ?? '';
}

function getPrimaryIPv6Subnet(routes: Route[]): string {
  const primaryRoute = routes.find((route) => !route.via && parseIPv6CIDR(route.target));
  return primaryRoute?.target ?? '';
}

function buildAutoAssignPoolFromSubnet(subnet: string): IpAssignmentPool | null {
  const parsed = parseIPv4CIDR(subnet);
  const prefix = getCIDRPrefix(subnet);
  if (!parsed || prefix === null) return null;

  if (prefix >= 31) {
    return null;
  }

  const size = 2 ** (32 - prefix);
  const start = parsed.network + 1;
  const end = parsed.network + size - 2;

  if (start > end) {
    return null;
  }

  return {
    ipRangeStart: formatIPv4Address(start >>> 0),
    ipRangeEnd: formatIPv4Address(end >>> 0),
  };
}

function isIPv4PoolCoveredByRoutes(pool: IpAssignmentPool, routes: Route[]): boolean {
  const start = parseIPv4Address(pool.ipRangeStart);
  const end = parseIPv4Address(pool.ipRangeEnd);
  if (start === null || end === null || start > end) {
    return false;
  }

  return routes.some((route) => {
    const parsedRoute = parseIPv4CIDR(route.target);
    if (!parsedRoute) return false;
    return (start & parsedRoute.mask) === parsedRoute.network && (end & parsedRoute.mask) === parsedRoute.network;
  });
}

function isPoolInsideSubnet(pool: IpAssignmentPool, subnet: string): boolean {
  const route = { target: subnet };
  return isIPv4PoolCoveredByRoutes(pool, [route]);
}

function isIPv6PoolInsideSubnet(pool: IpAssignmentPool, subnet: string): boolean {
  const parsedSubnet = parseIPv6CIDR(subnet);
  const start = parseIPv6Address(pool.ipRangeStart);
  const end = parseIPv6Address(pool.ipRangeEnd);
  if (!parsedSubnet || start === null || end === null || start > end) {
    return false;
  }

  const mask = parsedSubnet.prefix === 0 ? 0n : ((1n << BigInt(parsedSubnet.prefix)) - 1n) << BigInt(128 - parsedSubnet.prefix);
  return (start & mask) === parsedSubnet.network && (end & mask) === parsedSubnet.network;
}

function getIPv4RangeIssue(ipv4Subnet: string, ipPool: IpAssignmentPool | null): string | null {
  const trimmedSubnet = ipv4Subnet.trim();
  if (trimmedSubnet === '') {
    return 'IPv4 子网不能为空。';
  }

  const normalizedSubnet = normalizeIPv4CIDR(trimmedSubnet);
  if (!normalizedSubnet) {
    return 'IPv4 子网必须是有效的 CIDR，例如 10.22.2.0/24。';
  }

  const prefix = getCIDRPrefix(normalizedSubnet);
  if (prefix === null || prefix > 30) {
    return 'IPv4 子网前缀需要在 0 到 30 之间，才能用于成员地址分配。';
  }

  if (!ipPool) {
    return '必须提供一个有效的 IPv4 地址池范围。';
  }

  const start = ipPool.ipRangeStart.trim();
  const end = ipPool.ipRangeEnd.trim();

  if (start === '' || end === '') {
    return '自动分配地址池需要同时填写起始 IP 和结束 IP。';
  }

  const parsedStart = parseIPv4Address(start);
  const parsedEnd = parseIPv4Address(end);
  if (parsedStart === null || parsedEnd === null) {
    return '自动分配地址池不是有效的 IPv4 范围。';
  }

  if (parsedStart > parsedEnd) {
    return '自动分配地址池的起始 IP 不能大于结束 IP。';
  }

  if (!isPoolInsideSubnet(ipPool, normalizedSubnet)) {
    return '自动分配地址池必须完全落在 IPv4 子网内。';
  }

  return null;
}

function getIPv4ConfigurationIssues(ipv4Subnet: string, v4AssignMode: boolean, ipPools: IpAssignmentPool[]): string[] {
  const trimmedSubnet = ipv4Subnet.trim();
  if (trimmedSubnet === '') {
    return ['IPv4 子网不能为空。'];
  }

  const normalizedSubnet = normalizeIPv4CIDR(trimmedSubnet);
  if (!normalizedSubnet) {
    return ['IPv4 子网必须是有效的 CIDR，例如 10.22.2.0/24。'];
  }

  const prefix = getCIDRPrefix(normalizedSubnet);
  if (prefix === null || prefix > 30) {
    return ['IPv4 子网前缀需要在 0 到 30 之间，才能用于成员地址分配。'];
  }

  if (!v4AssignMode) {
    return [];
  }

  if (ipPools.length === 0) {
    return ['开启自动分配后，至少需要配置一个 IPv4 地址池范围。'];
  }

  const rangeIssues = ipPools.flatMap((ipPool, index) => {
    const issue = getIPv4RangeIssue(ipv4Subnet, ipPool);
    return issue ? [`地址池 ${index + 1}：${issue}`] : [];
  });

  return [...rangeIssues, ...getIPv4PoolOverlapIssues(ipPools)];
}

function areIpPoolsEqual(a: IpAssignmentPool[], b: IpAssignmentPool[]): boolean {
  if (a.length !== b.length) return false;
  return a.every((pool, index) =>
    pool.ipRangeStart === b[index]?.ipRangeStart &&
    pool.ipRangeEnd === b[index]?.ipRangeEnd
  );
}

function getIPv4PoolOverlapIssues(ipPools: IpAssignmentPool[]): string[] {
  const parsedPools = ipPools.map((pool, index) => ({
    index,
    start: parseIPv4Address(pool.ipRangeStart),
    end: parseIPv4Address(pool.ipRangeEnd),
  }));

  const issues: string[] = [];
  for (let i = 0; i < parsedPools.length; i += 1) {
    const current = parsedPools[i];
    if (current.start === null || current.end === null) continue;

    for (let j = i + 1; j < parsedPools.length; j += 1) {
      const next = parsedPools[j];
      if (next.start === null || next.end === null) continue;

      if (current.start <= next.end && next.start <= current.end) {
        issues.push(`地址池 ${current.index + 1} 与地址池 ${next.index + 1} 存在重叠。`);
      }
    }
  }

  return issues;
}

function getIPv6RangeIssue(ipv6Subnet: string, ipPool: IpAssignmentPool | null): string | null {
  const trimmedSubnet = ipv6Subnet.trim();
  if (trimmedSubnet === '') {
    return 'IPv6 子网不能为空。';
  }

  const normalizedSubnet = normalizeIPv6CIDR(trimmedSubnet);
  if (!normalizedSubnet) {
    return 'IPv6 子网必须是有效的 CIDR，例如 fd00::/48。';
  }

  const parsedSubnet = parseIPv6CIDR(normalizedSubnet);
  if (!parsedSubnet || parsedSubnet.prefix > 127) {
    return 'IPv6 子网前缀需要在 0 到 127 之间。';
  }

  if (!ipPool) {
    return '必须提供一个有效的 IPv6 地址池范围。';
  }

  const start = ipPool.ipRangeStart.trim();
  const end = ipPool.ipRangeEnd.trim();
  if (start === '' || end === '') {
    return '自定义 IPv6 地址池需要同时填写起始 IP 和结束 IP。';
  }

  const parsedStart = parseIPv6Address(start);
  const parsedEnd = parseIPv6Address(end);
  if (parsedStart === null || parsedEnd === null) {
    return '自定义 IPv6 地址池不是有效的 IPv6 范围。';
  }

  if (parsedStart > parsedEnd) {
    return '自定义 IPv6 地址池的起始 IP 不能大于结束 IP。';
  }

  if (!isIPv6PoolInsideSubnet(ipPool, normalizedSubnet)) {
    return '自定义 IPv6 地址池必须完全落在 IPv6 子网内。';
  }

  return null;
}

function getIPv6PoolOverlapIssues(ipPools: IpAssignmentPool[]): string[] {
  const parsedPools = ipPools.map((pool, index) => ({
    index,
    start: parseIPv6Address(pool.ipRangeStart),
    end: parseIPv6Address(pool.ipRangeEnd),
  }));

  const issues: string[] = [];
  for (let i = 0; i < parsedPools.length; i += 1) {
    const current = parsedPools[i];
    if (current.start === null || current.end === null) continue;

    for (let j = i + 1; j < parsedPools.length; j += 1) {
      const next = parsedPools[j];
      if (next.start === null || next.end === null) continue;

      if (current.start <= next.end && next.start <= current.end) {
        issues.push(`IPv6 地址池 ${current.index + 1} 与地址池 ${next.index + 1} 存在重叠。`);
      }
    }
  }

  return issues;
}

function getIPv6ConfigurationIssues(ipv6Subnet: string, customAssign: boolean, ipv6Pools: IpAssignmentPool[]): string[] {
  if (!customAssign) return [];

  if (ipv6Pools.length === 0) {
    return ['开启自定义 IPv6 范围后，至少需要配置一个 IPv6 地址池范围。'];
  }

  const rangeIssues = ipv6Pools.flatMap((pool, index) => {
    const issue = getIPv6RangeIssue(ipv6Subnet, pool);
    return issue ? [`IPv6 地址池 ${index + 1}：${issue}`] : [];
  });

  return [...rangeIssues, ...getIPv6PoolOverlapIssues(ipv6Pools)];
}

function generateRandomIPv4Subnet(): string {
  return `10.${Math.floor(Math.random() * 256)}.${Math.floor(Math.random() * 256)}.0/24`;
}

function formatNetworkMember(member: ApiMember): NetworkMemberDevice {
  const memberID = member.id || member.nodeId || '';
  const clientVersion = member.clientVersion || (
    member.vMajor !== undefined && member.vMinor !== undefined && member.vRev !== undefined &&
    member.vMajor >= 0 && member.vMinor >= 0 && member.vRev >= 0
      ? `${member.vMajor}.${member.vMinor}.${member.vRev}`
      : 'unknown'
  );
  return {
    id: memberID,
    name: member.name || member.nodeId || member.id || '未命名设备',
    description: member.description || '',
    authorized: member.config?.authorized ?? member.authorized ?? false,
    ipAssignments: member.config?.ipAssignments ?? member.ipAssignments ?? [],
    clientVersion,
    lastSeenLabel: member.lastSeen ? new Date(member.lastSeen).toLocaleString() : '未知',
    activeBridge: member.config?.activeBridge ?? member.activeBridge ?? false,
    noAutoAssignIps: member.config?.noAutoAssignIps ?? member.noAutoAssignIps ?? false,
  };
}

function NetworkDetail() {
  const { id } = useParams<{ id: string }>();
  const [network, setNetwork] = useState<Network | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [saving, setSaving] = useState<boolean>(false);
  const [error, setError] = useState<string>('');
  const [activeTab, setActiveTab] = useState<number>(0);
  const [editingName, setEditingName] = useState<string>('');
  const [editingDescription, setEditingDescription] = useState<string>('');
  const [v4AssignMode, setV4AssignMode] = useState<boolean>(false);
  const [ipv6Subnet, setIpv6Subnet] = useState<string>('');
  const [ipv6CustomAssign, setIpv6CustomAssign] = useState<boolean>(false);
  const [ipv6Rfc4193, setIpv6Rfc4193] = useState<boolean>(false);
  const [ipv6Plane6, setIpv6Plane6] = useState<boolean>(false);
  const [ipv6PoolStartDraft, setIpv6PoolStartDraft] = useState<string>('');
  const [ipv6PoolEndDraft, setIpv6PoolEndDraft] = useState<string>('');
  const [ipv6Pools, setIpv6Pools] = useState<IpAssignmentPool[]>([]);
  const [multicastLimit, setMulticastLimit] = useState<number>(32);
  const [enableBroadcast, setEnableBroadcast] = useState<boolean>(true);
  const [ipv4Subnet, setIpv4Subnet] = useState<string>('');
  const [ipv4PoolStartDraft, setIpv4PoolStartDraft] = useState<string>('');
  const [ipv4PoolEndDraft, setIpv4PoolEndDraft] = useState<string>('');
  const [ipv4Pools, setIpv4Pools] = useState<IpAssignmentPool[]>([]);
  const [routes, setRoutes] = useState<Route[]>([]);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);
  const [snackbar, setSnackbar] = useState<{ open: boolean; message: string; severity: 'success' | 'error' | 'info' }>({ open: false, message: '', severity: 'success' });
  const [basicInfoUnsaved, setBasicInfoUnsaved] = useState<boolean>(false);
  const [ipv4Unsaved, setIpv4Unsaved] = useState<boolean>(false);
  const [ipv6Unsaved, setIpv6Unsaved] = useState<boolean>(false);
  const [multicastUnsaved, setMulticastUnsaved] = useState<boolean>(false);
  const [memberSearchTerm, setMemberSearchTerm] = useState<string>('');
  const [memberMenuAnchor, setMemberMenuAnchor] = useState<null | HTMLElement>(null);
  const [selectedMember, setSelectedMember] = useState<NetworkMemberDevice | null>(null);
  const [memberDialogOpen, setMemberDialogOpen] = useState<boolean>(false);
  const [memberDeleteDialogOpen, setMemberDeleteDialogOpen] = useState<boolean>(false);
  const [memberForm, setMemberForm] = useState<{ name: string; authorized: boolean; activeBridge: boolean; noAutoAssignIps: boolean; ipAssignments: string[] }>({
    name: '',
    authorized: false,
    activeBridge: false,
    noAutoAssignIps: false,
    ipAssignments: [''],
  });
  const [hidePendingBanner, setHidePendingBanner] = useState<boolean>(false);

  useEffect(() => {
    void fetchNetworkDetail();
  }, [id]);

  useEffect(() => {
    if (network) {
      setBasicInfoUnsaved(editingName !== (network.name || '') || editingDescription !== (network.db_description || network.description || ''));
    }
  }, [editingName, editingDescription, network]);

  useEffect(() => {
    if (network) {
      const subnetChanged = ipv4Subnet !== getPrimaryIPv4Subnet(network.config.routes || []);
      const poolChanged = !areIpPoolsEqual(ipv4Pools, network.config.ipAssignmentPools || []);
      const v4Changed = v4AssignMode !== (network.config.v4AssignMode?.zt ?? false);
      setIpv4Unsaved(subnetChanged || poolChanged || v4Changed);
    }
  }, [ipv4Subnet, ipv4Pools, v4AssignMode, network]);

  useEffect(() => {
    if (network) {
      const subnetChanged = ipv6Subnet !== getPrimaryIPv6Subnet(network.config.routes || []);
      const customChanged = ipv6CustomAssign !== (network.config.v6AssignMode?.zt ?? false);
      const rfcChanged = ipv6Rfc4193 !== (network.config.v6AssignMode?.rfc4193 ?? false);
      const planeChanged = ipv6Plane6 !== (network.config.v6AssignMode?.['6plane'] ?? false);
      const poolChanged = !areIpPoolsEqual(ipv6Pools, (network.config.ipAssignmentPools || []).filter((pool) => parseIPv6Address(pool.ipRangeStart) !== null));
      setIpv6Unsaved(subnetChanged || customChanged || rfcChanged || planeChanged || poolChanged);
    }
  }, [ipv6Subnet, ipv6CustomAssign, ipv6Rfc4193, ipv6Plane6, ipv6Pools, network]);

  useEffect(() => {
    if (network) {
      const multicastChanged = multicastLimit !== (network.config.multicastLimit ?? 32);
      const broadcastChanged = enableBroadcast !== (network.config.enableBroadcast ?? true);
      setMulticastUnsaved(multicastChanged || broadcastChanged);
    }
  }, [multicastLimit, enableBroadcast, network]);

  const fetchNetworkDetail = async () => {
    setLoading(true);
    try {
      if (!id) {
        throw new Error('网络ID不能为空');
      }
      const response = await networkAPI.getNetwork(id);
      const networkData = response.data;
      if (networkData) {
        networkData.config = { ...defaultNetworkConfig, ...networkData.config };
        if (networkData.config.v4AssignMode) {
          networkData.config.v4AssignMode = { ...defaultNetworkConfig.v4AssignMode, ...networkData.config.v4AssignMode };
        }
        if (networkData.config.v6AssignMode) {
          networkData.config.v6AssignMode = { ...defaultNetworkConfig.v6AssignMode, ...networkData.config.v6AssignMode };
        }
        setNetwork(networkData);
        setEditingName(networkData.name || '');
        setEditingDescription(networkData.db_description || networkData.description || '');
        setV4AssignMode(networkData.config.v4AssignMode?.zt ?? false);
        setIpv6Subnet(getPrimaryIPv6Subnet(networkData.config.routes || []));
        setIpv6CustomAssign(networkData.config.v6AssignMode?.zt ?? false);
        setIpv6Rfc4193(networkData.config.v6AssignMode?.rfc4193 ?? false);
        setIpv6Plane6(networkData.config.v6AssignMode?.['6plane'] ?? false);
        const loadedIPv6Pools = (networkData.config.ipAssignmentPools || []).filter((pool) => parseIPv6Address(pool.ipRangeStart) !== null);
        setIpv6Pools(loadedIPv6Pools);
        setIpv6PoolStartDraft(loadedIPv6Pools[0]?.ipRangeStart ?? '');
        setIpv6PoolEndDraft(loadedIPv6Pools[0]?.ipRangeEnd ?? '');
        setMulticastLimit(networkData.config.multicastLimit ?? 32);
        setEnableBroadcast(networkData.config.enableBroadcast ?? true);
        const primarySubnet = getPrimaryIPv4Subnet(networkData.config.routes || []);
        setIpv4Subnet(primarySubnet);
        const loadedPools = networkData.config.ipAssignmentPools || [];
        setIpv4Pools(loadedPools);
        setIpv4PoolStartDraft(loadedPools[0]?.ipRangeStart ?? '');
        setIpv4PoolEndDraft(loadedPools[0]?.ipRangeEnd ?? '');
        setRoutes(networkData.config.routes || []);
        setHidePendingBanner(false);
      }
    } catch (err: unknown) {
      setError(getErrorMessage(err, '获取网络详情失败'));
      console.error('Fetch network detail error:', err);
    } finally {
      setLoading(false);
    }
  };

  const showSnackbar = (message: string, severity: 'success' | 'error' | 'info' = 'success') => {
    setSnackbar({ open: true, message, severity });
  };

  const memberDevices = (network?.members || []).map(formatNetworkMember);
  const pendingMembers = memberDevices.filter((member) => !member.authorized);
  const authorizedMembers = memberDevices.filter((member) => member.authorized);
  const ipv4RangeDraftIssue = getIPv4RangeIssue(ipv4Subnet, {
    ipRangeStart: ipv4PoolStartDraft,
    ipRangeEnd: ipv4PoolEndDraft,
  });
  const ipv4ConfigurationIssues = getIPv4ConfigurationIssues(ipv4Subnet, v4AssignMode, ipv4Pools);
  const ipv6RangeDraftIssue = getIPv6RangeIssue(ipv6Subnet, {
    ipRangeStart: ipv6PoolStartDraft,
    ipRangeEnd: ipv6PoolEndDraft,
  });
  const ipv6ConfigurationIssues = getIPv6ConfigurationIssues(ipv6Subnet, ipv6CustomAssign, ipv6Pools);
  const filteredMembers = memberDevices.filter((member) => {
    const query = memberSearchTerm.trim().toLowerCase();
    if (query === '') return true;
    return member.name.toLowerCase().includes(query) || member.id.toLowerCase().includes(query);
  });

  const closeMemberMenu = () => {
    setMemberMenuAnchor(null);
  };

  const handleOpenMemberMenu = (event: React.MouseEvent<HTMLElement>, member: NetworkMemberDevice) => {
    setMemberMenuAnchor(event.currentTarget);
    setSelectedMember(member);
  };

  const handleOpenEditMember = () => {
    if (!selectedMember) return;
    setMemberForm({
      name: selectedMember.name,
      authorized: selectedMember.authorized,
      activeBridge: selectedMember.activeBridge,
      noAutoAssignIps: selectedMember.noAutoAssignIps,
      ipAssignments: selectedMember.ipAssignments.length > 0 ? [...selectedMember.ipAssignments] : [''],
    });
    setMemberDialogOpen(true);
    closeMemberMenu();
  };

  const handleUpdateMemberStatus = async (member: NetworkMemberDevice, authorized: boolean) => {
    if (!id) return;
    setSaving(true);
    try {
      await memberAPI.updateMember(id, member.id, { authorized });
      showSnackbar(authorized ? '设备授权成功' : '设备已拒绝');
      await fetchNetworkDetail();
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, authorized ? '设备授权失败' : '设备拒绝失败'), 'error');
    } finally {
      setSaving(false);
      closeMemberMenu();
    }
  };

  const handleSaveMember = async () => {
    if (!id || !selectedMember) return;
    setSaving(true);
    try {
      await memberAPI.updateMember(id, selectedMember.id, {
        name: memberForm.name,
        authorized: memberForm.authorized,
        activeBridge: memberForm.activeBridge,
        noAutoAssignIps: memberForm.noAutoAssignIps,
        ipAssignments: memberForm.ipAssignments.map((ip) => ip.trim()).filter((ip) => ip !== ''),
      });
      showSnackbar('成员信息更新成功');
      setMemberDialogOpen(false);
      await fetchNetworkDetail();
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '更新成员失败'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteMember = async () => {
    if (!id || !selectedMember) return;
    setSaving(true);
    try {
      await memberAPI.deleteMember(id, selectedMember.id);
      showSnackbar('成员移除成功');
      setMemberDeleteDialogOpen(false);
      closeMemberMenu();
      await fetchNetworkDetail();
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '移除成员失败'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleMemberIpChange = (index: number, value: string) => {
    setMemberForm((prev) => {
      const next = [...prev.ipAssignments];
      next[index] = value;
      return { ...prev, ipAssignments: next };
    });
  };

  const handleAddMemberIP = () => {
    setMemberForm((prev) => ({ ...prev, ipAssignments: [...prev.ipAssignments, ''] }));
  };

  const handleRemoveMemberIP = (index: number) => {
    setMemberForm((prev) => {
      const next = prev.ipAssignments.filter((_, currentIndex) => currentIndex !== index);
      return { ...prev, ipAssignments: next.length > 0 ? next : [''] };
    });
  };

  const handleCopyMemberID = async () => {
    if (!selectedMember) return;
    try {
      await navigator.clipboard.writeText(selectedMember.id);
      showSnackbar('设备 ID 已复制');
    } catch {
      showSnackbar('复制设备 ID 失败', 'error');
    }
  };

  const handleSaveMetadata = async () => {
    if (!id || !network) return;
    setSaving(true);
    try {
      const metadataData: NetworkMetadataUpdateRequest = {
        name: editingName,
        description: editingDescription
      };
      await networkAPI.updateNetworkMetadata(id, metadataData);
      showSnackbar('名称和描述保存成功');
      setTimeout(() => { void fetchNetworkDetail(); }, 1000);
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleResetMetadata = () => {
    if (!network) return;
    setEditingName(network.name || '');
    setEditingDescription(network.db_description || network.description || '');
  };

  const handleDeleteNetwork = async () => {
    if (!id) return;
    setSaving(true);
    try {
      await networkAPI.deleteNetwork(id);
      showSnackbar('网络删除成功', 'success');
      setTimeout(() => {
        window.location.href = '/networks';
      }, 1000);
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '删除失败'), 'error');
    } finally {
      setSaving(false);
      setDeleteDialogOpen(false);
    }
  };

  const handleSaveIPv4 = async () => {
    if (!id) return;
    if (ipv4ConfigurationIssues.length > 0) {
      showSnackbar(ipv4ConfigurationIssues[0], 'error');
      return;
    }
    setSaving(true);
    try {
      const normalizedSubnet = normalizeIPv4CIDR(ipv4Subnet);
      const currentPrimaryIPv4Subnet = getPrimaryIPv4Subnet(routes);
      const currentPrimaryIPv6Subnet = getPrimaryIPv6Subnet(routes);
      const updatedRoutes = [
        ...(normalizedSubnet ? [{ target: normalizedSubnet }] : []),
        ...(ipv6CustomAssign && normalizeIPv6CIDR(ipv6Subnet) ? [{ target: normalizeIPv6CIDR(ipv6Subnet) as string }] : []),
        ...routes.filter((route) =>
          route.via ||
          (route.target !== currentPrimaryIPv4Subnet && route.target !== currentPrimaryIPv6Subnet)
        ),
      ];
      const updatedPools = [
        ...(v4AssignMode ? ipv4Pools.map((pool) => ({
          ipRangeStart: pool.ipRangeStart.trim(),
          ipRangeEnd: pool.ipRangeEnd.trim(),
        })) : []),
        ...(ipv6CustomAssign ? ipv6Pools.map((pool) => ({
          ipRangeStart: pool.ipRangeStart.trim(),
          ipRangeEnd: pool.ipRangeEnd.trim(),
        })) : []),
      ];

      await networkAPI.updateNetwork(id, {
        ipAssignmentPools: updatedPools,
        routes: updatedRoutes,
        v4AssignMode: { zt: v4AssignMode }
      });
      showSnackbar('IPv4配置保存成功');
      setTimeout(() => { void fetchNetworkDetail(); }, 1000);
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleResetIPv4 = () => {
    if (!network) return;
    const primarySubnet = getPrimaryIPv4Subnet(network.config.routes || []);
    const loadedPools = network.config.ipAssignmentPools || [];
    setIpv4Subnet(primarySubnet);
    setV4AssignMode(network.config.v4AssignMode?.zt ?? false);
    setIpv4Pools(loadedPools);
    setIpv4PoolStartDraft(loadedPools[0]?.ipRangeStart ?? '');
    setIpv4PoolEndDraft(loadedPools[0]?.ipRangeEnd ?? '');
  };

  const handleSaveIPv6 = async () => {
    if (!id) return;
    if (ipv6ConfigurationIssues.length > 0) {
      showSnackbar(ipv6ConfigurationIssues[0], 'error');
      return;
    }
    setSaving(true);
    try {
      const currentPrimaryIPv4Subnet = getPrimaryIPv4Subnet(routes);
      const currentPrimaryIPv6Subnet = getPrimaryIPv6Subnet(routes);
      const normalizedIPv4Subnet = normalizeIPv4CIDR(ipv4Subnet);
      const normalizedIPv6Subnet = normalizeIPv6CIDR(ipv6Subnet);
      const updatedRoutes = [
        ...(normalizedIPv4Subnet ? [{ target: normalizedIPv4Subnet }] : []),
        ...(ipv6CustomAssign && normalizedIPv6Subnet ? [{ target: normalizedIPv6Subnet }] : []),
        ...routes.filter((route) =>
          route.via ||
          (route.target !== currentPrimaryIPv4Subnet && route.target !== currentPrimaryIPv6Subnet)
        ),
      ];
      const updatedPools = [
        ...(v4AssignMode ? ipv4Pools.map((pool) => ({
          ipRangeStart: pool.ipRangeStart.trim(),
          ipRangeEnd: pool.ipRangeEnd.trim(),
        })) : []),
        ...(ipv6CustomAssign ? ipv6Pools.map((pool) => ({
          ipRangeStart: pool.ipRangeStart.trim(),
          ipRangeEnd: pool.ipRangeEnd.trim(),
        })) : []),
      ];
      await networkAPI.updateNetwork(id, {
        routes: updatedRoutes,
        ipAssignmentPools: updatedPools,
        v6AssignMode: {
          zt: ipv6CustomAssign,
          '6plane': ipv6Plane6,
          rfc4193: ipv6Rfc4193,
        }
      });
      showSnackbar('IPv6配置保存成功');
      setTimeout(() => { void fetchNetworkDetail(); }, 1000);
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleResetIPv6 = () => {
    if (!network) return;
    setIpv6Subnet(getPrimaryIPv6Subnet(network.config.routes || []));
    setIpv6CustomAssign(network.config.v6AssignMode?.zt ?? false);
    setIpv6Rfc4193(network.config.v6AssignMode?.rfc4193 ?? false);
    setIpv6Plane6(network.config.v6AssignMode?.['6plane'] ?? false);
    const loadedIPv6Pools = (network.config.ipAssignmentPools || []).filter((pool) => parseIPv6Address(pool.ipRangeStart) !== null);
    setIpv6Pools(loadedIPv6Pools);
    setIpv6PoolStartDraft(loadedIPv6Pools[0]?.ipRangeStart ?? '');
    setIpv6PoolEndDraft(loadedIPv6Pools[0]?.ipRangeEnd ?? '');
  };

  const handleSaveMulticast = async () => {
    if (!id) return;
    setSaving(true);
    try {
      await networkAPI.updateNetwork(id, {
        multicastLimit: multicastLimit,
        enableBroadcast: enableBroadcast
      });
      showSnackbar('多播设置保存成功');
      setTimeout(() => { void fetchNetworkDetail(); }, 1000);
    } catch (err: unknown) {
      showSnackbar(getErrorMessage(err, '保存失败'), 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleResetMulticast = () => {
    if (!network) return;
    setMulticastLimit(network.config.multicastLimit ?? 32);
    setEnableBroadcast(network.config.enableBroadcast ?? true);
  };

  const handleGenerateSubnetAndRange = () => {
    const subnet = generateRandomIPv4Subnet();
    const pool = buildAutoAssignPoolFromSubnet(subnet);
    setIpv4Subnet(subnet);
    if (pool) {
      setIpv4PoolStartDraft(pool.ipRangeStart);
      setIpv4PoolEndDraft(pool.ipRangeEnd);
    }
  };

  const handleToggleV4AssignMode = (enabled: boolean) => {
    setV4AssignMode(enabled);
    if (enabled && ipv4Pools.length === 0 && (!ipv4PoolStartDraft || !ipv4PoolEndDraft)) {
      const pool = buildAutoAssignPoolFromSubnet(ipv4Subnet);
      if (pool) {
        setIpv4PoolStartDraft(pool.ipRangeStart);
        setIpv4PoolEndDraft(pool.ipRangeEnd);
      }
    }
    if (!enabled) {
      setIpv4PoolStartDraft('');
      setIpv4PoolEndDraft('');
      setIpv4Pools([]);
    }
  };

  const handleSubnetChange = (value: string) => {
    setIpv4Subnet(value);
    if (v4AssignMode && ipv4Pools.length === 0) {
      const pool = buildAutoAssignPoolFromSubnet(value);
      if (pool) {
        setIpv4PoolStartDraft(pool.ipRangeStart);
        setIpv4PoolEndDraft(pool.ipRangeEnd);
      }
    }
  };

  const handleSetIPv4Range = () => {
    const nextPool = {
      ipRangeStart: ipv4PoolStartDraft.trim(),
      ipRangeEnd: ipv4PoolEndDraft.trim(),
    };
    const issue = getIPv4RangeIssue(ipv4Subnet, nextPool);
    if (issue) {
      showSnackbar(issue, 'error');
      return;
    }

    if (ipv4Pools.some((pool) => pool.ipRangeStart === nextPool.ipRangeStart && pool.ipRangeEnd === nextPool.ipRangeEnd)) {
      showSnackbar('该地址池范围已存在', 'info');
      return;
    }

    const overlapIssues = getIPv4PoolOverlapIssues([...ipv4Pools, nextPool]);
    if (overlapIssues.length > 0) {
      showSnackbar(overlapIssues[0], 'error');
      return;
    }

    setIpv4Pools((prev) => [...prev, nextPool]);
    setIpv4PoolStartDraft('');
    setIpv4PoolEndDraft('');
  };

  const handleRemoveIPv4Range = (index: number) => {
    setIpv4Pools((prev) => prev.filter((_, currentIndex) => currentIndex !== index));
  };

  const handleSetIPv6Range = () => {
    const nextPool = {
      ipRangeStart: ipv6PoolStartDraft.trim(),
      ipRangeEnd: ipv6PoolEndDraft.trim(),
    };
    const issue = getIPv6RangeIssue(ipv6Subnet, nextPool);
    if (issue) {
      showSnackbar(issue, 'error');
      return;
    }

    if (ipv6Pools.some((pool) => pool.ipRangeStart === nextPool.ipRangeStart && pool.ipRangeEnd === nextPool.ipRangeEnd)) {
      showSnackbar('该 IPv6 地址池范围已存在', 'info');
      return;
    }

    const overlapIssues = getIPv6PoolOverlapIssues([...ipv6Pools, nextPool]);
    if (overlapIssues.length > 0) {
      showSnackbar(overlapIssues[0], 'error');
      return;
    }

    setIpv6Pools((prev) => [...prev, nextPool]);
    setIpv6PoolStartDraft('');
    setIpv6PoolEndDraft('');
  };

  const handleRemoveIPv6Range = (index: number) => {
    setIpv6Pools((prev) => prev.filter((_, currentIndex) => currentIndex !== index));
  };

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setActiveTab(newValue);
  };

  const handleCloseSnackbar = () => {
    setSnackbar({ ...snackbar, open: false });
  };

  if (error || !id) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="error" sx={{ mb: 3 }}>
          {error || '网络不存在'}
        </Alert>
        <Button 
          variant="contained" 
          component={Link} 
          to="/networks"
        >
          返回网络列表
        </Button>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      {/* 标题栏始终显示 */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <IconButton 
            component={Link} 
            to="/networks"
            size="large"
          >
            <ArrowBack />
          </IconButton>
          <Typography variant="h4" component="h1">
            {network?.name}
          </Typography>
        </Box>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Typography variant="body1" color="text.secondary">
            网络ID: {network?.id}
          </Typography>
          <IconButton size="small">
            <ContentCopy />
          </IconButton>
        </Box>
      </Box>

      {/* 导航标签页始终显示 */}
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 3 }}>
        <Tabs value={activeTab} onChange={handleTabChange} aria-label="network tabs">
          <Tab label="成员设备" />
          <Tab label="设置" />
        </Tabs>
      </Box>

      {/* 加载状态显示 */}
      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 10 }}>
          <CircularProgress />
        </Box>
      ) : (
        // 内容区域
        <>
          {/* 成员设备选项卡内容 */}
          {activeTab === 0 && (
            <>
              {pendingMembers.length > 0 && !hidePendingBanner && (
                <Alert
                  severity="warning"
                  sx={{ mb: 3 }}
                  action={(
                    <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
                      <Button
                        color="inherit"
                        variant="outlined"
                        size="small"
                        disabled={saving}
                        onClick={() => { void handleUpdateMemberStatus(pendingMembers[0], true); }}
                      >
                        授权第一个待审批成员
                      </Button>
                      <Button
                        color="inherit"
                        variant="text"
                        size="small"
                        disabled={saving}
                        onClick={() => { void handleUpdateMemberStatus(pendingMembers[0], false); }}
                      >
                        拒绝第一个待审批成员
                      </Button>
                      <Button color="inherit" size="small" onClick={() => setHidePendingBanner(true)}>
                        关闭
                      </Button>
                    </Stack>
                  )}
                >
                  当前有 {pendingMembers.length} 个待授权设备。你可以在此快速审批，也可以在下方成员列表中逐个处理。
                </Alert>
              )}

              {/* 统计卡片区域 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Grid container spacing={3}>
                  <Grid size={{ xs: 12, sm: 4 }}>
                    <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
                      <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                        <Typography variant="h6" color="text.secondary" gutterBottom>
                          设备总数
                        </Typography>
                        <Typography variant="h4">
                          {memberDevices.length}
                        </Typography>
                      </CardContent>
                    </Card>
                  </Grid>
                  <Grid size={{ xs: 12, sm: 4 }}>
                    <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
                      <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                        <Typography variant="h6" color="text.secondary" gutterBottom>
                          已授权设备
                        </Typography>
                        <Typography variant="h4">
                          {authorizedMembers.length}
                        </Typography>
                      </CardContent>
                    </Card>
                  </Grid>
                  <Grid size={{ xs: 12, sm: 4 }}>
                    <Card sx={{ height: '100%', backgroundColor: '#2c3e50', display: 'flex', flexDirection: 'column' }}>
                      <CardContent sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                        <Typography variant="h6" color="text.secondary" gutterBottom>
                          待授权设备
                        </Typography>
                        <Typography variant="h4">
                          {pendingMembers.length}
                        </Typography>
                      </CardContent>
                    </Card>
                  </Grid>
                </Grid>
              </Paper>

              {/* 成员设备部分 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
                  <Typography variant="h5">
                    成员设备
                  </Typography>
                  <TextField
                    size="small"
                    placeholder="搜索设备名称或设备 ID"
                    value={memberSearchTerm}
                    onChange={(event) => setMemberSearchTerm(event.target.value)}
                    sx={{ width: { xs: '100%', sm: 280 } }}
                  />
                </Box>
                <TableContainer>
                  <Table>
                    <TableHead>
                      <TableRow>
                        <TableCell>设备 ID</TableCell>
                        <TableCell>名称</TableCell>
                        <TableCell>状态</TableCell>
                        <TableCell>Managed IPs</TableCell>
                        <TableCell>ZT 版本</TableCell>
                        <TableCell>最后活动</TableCell>
                        <TableCell align="right">操作</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {filteredMembers.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={7} align="center" sx={{ py: 5, color: 'text.secondary' }}>
                            {memberSearchTerm ? '没有找到匹配的成员设备' : '暂无设备连接'}
                          </TableCell>
                        </TableRow>
                      ) : (
                        filteredMembers.map((member) => (
                          <TableRow key={member.id} hover>
                            <TableCell>{member.id}</TableCell>
                            <TableCell>{member.name}</TableCell>
                            <TableCell>
                              <Chip
                                label={member.authorized ? '已授权' : '待授权'}
                                color={member.authorized ? 'success' : 'warning'}
                                variant={member.authorized ? 'filled' : 'outlined'}
                                size="small"
                              />
                            </TableCell>
                            <TableCell>{member.ipAssignments.length > 0 ? member.ipAssignments.join(', ') : '-'}</TableCell>
                            <TableCell>{member.clientVersion}</TableCell>
                            <TableCell>{member.lastSeenLabel}</TableCell>
                            <TableCell align="right">
                              <IconButton onClick={(event) => handleOpenMemberMenu(event, member)}>
                                <MoreHoriz />
                              </IconButton>
                            </TableCell>
                          </TableRow>
                        ))
                      )}
                    </TableBody>
                  </Table>
                </TableContainer>
              </Paper>
            </>
          )}

          {/* 设置选项卡内容 */}
          {activeTab === 1 && (
            <>
              {/* 网络基本信息 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4, border: basicInfoUnsaved ? '2px solid' : 'none', borderColor: 'warning.main' }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h5">
                    网络基本信息
                  </Typography>
                  {basicInfoUnsaved && (
                    <Typography variant="body2" color="warning.main">
                      未保存
                    </Typography>
                  )}
                </Box>
                <Grid container spacing={3}>
                  <Grid size={{ xs: 12 }}>
                    <TextField
                      fullWidth
                      label="网络名称"
                      variant="outlined"
                      value={editingName}
                      onChange={(e) => setEditingName(e.target.value)}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12 }}>
                    <TextField
                      fullWidth
                      label="网络描述"
                      variant="outlined"
                      multiline
                      rows={3}
                      value={editingDescription}
                      onChange={(e) => setEditingDescription(e.target.value)}
                      sx={{ mb: 2 }}
                    />
                  </Grid>
                  <Grid size={{ xs: 12 }}>
                    <Box sx={{ display: 'flex', justifyContent: 'flex-start', gap: 2 }}>
                      <Button variant="outlined" onClick={handleResetMetadata} disabled={saving || !basicInfoUnsaved}>
                        重置更改
                      </Button>
                      <Button variant="contained" color="primary" onClick={() => { void handleSaveMetadata(); }} disabled={saving || !basicInfoUnsaved}>
                        保存
                      </Button>
                    </Box>
                  </Grid>
                </Grid>
              </Paper>
              
              {/* IPv4分配 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4, border: ipv4Unsaved ? '2px solid' : 'none', borderColor: 'warning.main' }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h5">
                    IPv4分配
                  </Typography>
                  {ipv4Unsaved && (
                    <Typography variant="body2" color="warning.main">
                      未保存
                    </Typography>
                  )}
                </Box>
                <Box sx={{ mb: 3 }}>
                  <Typography variant="body1" sx={{ mb: 2 }}>
                    默认会以一个 IPv4 子网作为网络边界。自动分配地址池必须完全落在该子网内。
                  </Typography>
                  <TextField
                    fullWidth
                    label="IPv4 子网"
                    placeholder="例如 10.22.2.0/24"
                    value={ipv4Subnet}
                    onChange={(e) => handleSubnetChange(e.target.value)}
                    sx={{ mb: 2 }}
                  />
                  <Box sx={{ display: 'flex', gap: 2, alignItems: 'center', flexWrap: 'wrap' }}>
                    <Button variant="outlined" onClick={handleGenerateSubnetAndRange}>
                      生成新的子网与范围
                    </Button>
                    <Typography variant="body2" color="text.secondary">
                      保存时会自动将该子网同步为网络的主托管路由。
                    </Typography>
                  </Box>
                </Box>

                <Paper variant="outlined" sx={{ p: 2.5, mb: 3, bgcolor: 'action.hover' }}>
                  <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 2, mb: v4AssignMode ? 2 : 0, flexWrap: 'wrap' }}>
                    <Box>
                      <Typography variant="h6">
                        自动分配 IPv4 地址给新成员设备
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        关闭后会收起自动分配范围，并在保存时清空 `ipAssignmentPools`。成员手动指定 IP 不受影响。
                      </Typography>
                    </Box>
                    <Switch checked={v4AssignMode} onChange={(e) => handleToggleV4AssignMode(e.target.checked)} />
                  </Box>

                  {v4AssignMode && (
                    <>
                      <Grid container spacing={2}>
                        <Grid size={{ xs: 12, md: 6 }}>
                          <TextField
                            fullWidth
                            label="起始 IPv4"
                            placeholder="例如 10.22.2.1"
                            value={ipv4PoolStartDraft}
                            onChange={(e) => setIpv4PoolStartDraft(e.target.value)}
                            error={Boolean(ipv4PoolStartDraft || ipv4PoolEndDraft) && Boolean(ipv4RangeDraftIssue)}
                            helperText={ipv4Subnet ? `必须落在 ${ipv4Subnet} 内` : '请先填写 IPv4 子网'}
                          />
                        </Grid>
                        <Grid size={{ xs: 12, md: 6 }}>
                          <TextField
                            fullWidth
                            label="结束 IPv4"
                            placeholder="例如 10.22.2.254"
                            value={ipv4PoolEndDraft}
                            onChange={(e) => setIpv4PoolEndDraft(e.target.value)}
                            error={Boolean(ipv4PoolStartDraft || ipv4PoolEndDraft) && Boolean(ipv4RangeDraftIssue)}
                            helperText={ipv4Subnet ? `必须落在 ${ipv4Subnet} 内` : '请先填写 IPv4 子网'}
                          />
                        </Grid>
                      </Grid>

                      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 2, flexWrap: 'wrap', mt: 2 }}>
                        <Box>
                          <Typography variant="subtitle1">
                            Active IPv4 Auto-Assign Range
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            使用 `Set Range` 将当前输入加入自动分配地址池列表。
                          </Typography>
                        </Box>
                        <Button variant="outlined" onClick={handleSetIPv4Range} disabled={Boolean(ipv4RangeDraftIssue)}>
                          Set Range
                        </Button>
                      </Box>

                      {Boolean(ipv4RangeDraftIssue) && Boolean(ipv4PoolStartDraft || ipv4PoolEndDraft) && (
                        <Alert severity="warning" sx={{ mt: 2 }}>
                          {ipv4RangeDraftIssue}
                        </Alert>
                      )}

                      {ipv4Pools.length > 0 ? (
                        <TableContainer component={Paper} variant="outlined" sx={{ mt: 2 }}>
                          <Table size="small">
                            <TableHead>
                              <TableRow>
                                <TableCell>Start</TableCell>
                                <TableCell>End</TableCell>
                                <TableCell align="right">Action</TableCell>
                              </TableRow>
                            </TableHead>
                            <TableBody>
                              {ipv4Pools.map((pool, index) => (
                                <TableRow key={`${pool.ipRangeStart}-${pool.ipRangeEnd}`}>
                                  <TableCell>{pool.ipRangeStart}</TableCell>
                                  <TableCell>{pool.ipRangeEnd}</TableCell>
                                  <TableCell align="right">
                                    <Button variant="outlined" color="error" size="small" onClick={() => handleRemoveIPv4Range(index)}>
                                      删除
                                    </Button>
                                  </TableCell>
                                </TableRow>
                              ))}
                            </TableBody>
                          </Table>
                        </TableContainer>
                      ) : (
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                          尚未配置自动分配地址池。
                        </Typography>
                      )}
                    </>
                  )}
                </Paper>

                {ipv4ConfigurationIssues.length > 0 && (
                  <Alert severity="warning" sx={{ mb: 3 }}>
                    {ipv4ConfigurationIssues.join(' ')}
                  </Alert>
                )}
                
                <Box sx={{ display: 'flex', justifyContent: 'flex-start', gap: 2 }}>
                  <Button variant="outlined" onClick={handleResetIPv4} disabled={saving || !ipv4Unsaved}>
                    重置更改
                  </Button>
                  <Button variant="contained" color="primary" onClick={() => { void handleSaveIPv4(); }} disabled={saving || !ipv4Unsaved}>
                    保存
                  </Button>
                </Box>
              </Paper>
              
              {/* IPv6分配 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4, border: ipv6Unsaved ? '2px solid' : 'none', borderColor: 'warning.main' }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h5">
                    IPv6分配
                  </Typography>
                  {ipv6Unsaved && (
                    <Typography variant="body2" color="warning.main">
                      未保存
                    </Typography>
                  )}
                </Box>
                
                <Box sx={{ mb: 3 }}>
                  <Typography variant="body1" sx={{ mb: 2 }}>
                    IPv6 默认不分配。只有自定义 IPv6 范围依赖手动填写子网；RFC4193 和 6PLANE 由控制器自动派生。
                  </Typography>
                  <TextField
                    fullWidth
                    label="IPv6 子网"
                    placeholder="例如 fd00::/48"
                    value={ipv6Subnet}
                    onChange={(e) => setIpv6Subnet(e.target.value)}
                  />
                </Box>

                <Paper variant="outlined" sx={{ p: 2.5, mb: 3, bgcolor: 'action.hover' }}>
                  <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 2, mb: ipv6CustomAssign ? 2 : 0, flexWrap: 'wrap' }}>
                    <Box>
                      <Typography variant="h6">
                        Assign from Custom IPv6 Range
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        需要先填写合法的 IPv6 子网，然后才能配置自定义 IPv6 地址池。
                      </Typography>
                    </Box>
                    <Switch
                      checked={ipv6CustomAssign}
                      disabled={!normalizeIPv6CIDR(ipv6Subnet)}
                      onChange={(e) => {
                        const enabled = e.target.checked;
                        setIpv6CustomAssign(enabled);
                        if (!enabled) {
                          setIpv6PoolStartDraft('');
                          setIpv6PoolEndDraft('');
                          setIpv6Pools([]);
                        }
                      }}
                    />
                  </Box>

                  {ipv6CustomAssign && (
                    <>
                      <Grid container spacing={2}>
                        <Grid size={{ xs: 12, md: 6 }}>
                          <TextField
                            fullWidth
                            label="起始 IPv6"
                            placeholder="例如 fd00::1000"
                            value={ipv6PoolStartDraft}
                            onChange={(e) => setIpv6PoolStartDraft(e.target.value)}
                            error={Boolean(ipv6PoolStartDraft || ipv6PoolEndDraft) && Boolean(ipv6RangeDraftIssue)}
                            helperText={ipv6Subnet ? `必须落在 ${ipv6Subnet} 内` : '请先填写 IPv6 子网'}
                          />
                        </Grid>
                        <Grid size={{ xs: 12, md: 6 }}>
                          <TextField
                            fullWidth
                            label="结束 IPv6"
                            placeholder="例如 fd00::1fff"
                            value={ipv6PoolEndDraft}
                            onChange={(e) => setIpv6PoolEndDraft(e.target.value)}
                            error={Boolean(ipv6PoolStartDraft || ipv6PoolEndDraft) && Boolean(ipv6RangeDraftIssue)}
                            helperText={ipv6Subnet ? `必须落在 ${ipv6Subnet} 内` : '请先填写 IPv6 子网'}
                          />
                        </Grid>
                      </Grid>

                      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 2, flexWrap: 'wrap', mt: 2 }}>
                        <Box>
                          <Typography variant="subtitle1">
                            Active IPv6 Auto-Assign Range
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            使用 `Set Range` 将当前输入加入 IPv6 地址池列表。
                          </Typography>
                        </Box>
                        <Button variant="outlined" onClick={handleSetIPv6Range} disabled={Boolean(ipv6RangeDraftIssue)}>
                          Set Range
                        </Button>
                      </Box>

                      {Boolean(ipv6RangeDraftIssue) && Boolean(ipv6PoolStartDraft || ipv6PoolEndDraft) && (
                        <Alert severity="warning" sx={{ mt: 2 }}>
                          {ipv6RangeDraftIssue}
                        </Alert>
                      )}

                      {ipv6Pools.length > 0 ? (
                        <TableContainer component={Paper} variant="outlined" sx={{ mt: 2 }}>
                          <Table size="small">
                            <TableHead>
                              <TableRow>
                                <TableCell>Start</TableCell>
                                <TableCell>End</TableCell>
                                <TableCell align="right">Action</TableCell>
                              </TableRow>
                            </TableHead>
                            <TableBody>
                              {ipv6Pools.map((pool, index) => (
                                <TableRow key={`${pool.ipRangeStart}-${pool.ipRangeEnd}`}>
                                  <TableCell>{pool.ipRangeStart}</TableCell>
                                  <TableCell>{pool.ipRangeEnd}</TableCell>
                                  <TableCell align="right">
                                    <Button variant="outlined" color="error" size="small" onClick={() => handleRemoveIPv6Range(index)}>
                                      删除
                                    </Button>
                                  </TableCell>
                                </TableRow>
                              ))}
                            </TableBody>
                          </Table>
                        </TableContainer>
                      ) : (
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                          尚未配置 IPv6 地址池。
                        </Typography>
                      )}
                    </>
                  )}
                </Paper>

                <Paper variant="outlined" sx={{ p: 2.5, mb: 3, bgcolor: 'action.hover' }}>
                  <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 2, flexWrap: 'wrap' }}>
                    <Box>
                      <Typography variant="h6">
                        Assign RFC4193 Unique Local Addresses (/128 per device)
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        自动为每个设备分配稳定的唯一本地 IPv6 地址。无需额外配置。
                      </Typography>
                    </Box>
                    <Switch checked={ipv6Rfc4193} onChange={(e) => setIpv6Rfc4193(e.target.checked)} />
                  </Box>
                </Paper>

                <Paper variant="outlined" sx={{ p: 2.5, mb: 3, bgcolor: 'action.hover' }}>
                  <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 2, flexWrap: 'wrap' }}>
                    <Box>
                      <Typography variant="h6">
                        Assign 6PLANE Routed Addresses (/80 per device)
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        为每个成员分配由节点 ID 派生的可路由 IPv6 地址。无需额外配置。
                      </Typography>
                    </Box>
                    <Switch checked={ipv6Plane6} onChange={(e) => setIpv6Plane6(e.target.checked)} />
                  </Box>
                </Paper>

                {ipv6ConfigurationIssues.length > 0 && (
                  <Alert severity="warning" sx={{ mb: 3 }}>
                    {ipv6ConfigurationIssues.join(' ')}
                  </Alert>
                )}

                <Box sx={{ display: 'flex', justifyContent: 'flex-start', gap: 2 }}>
                  <Button variant="outlined" onClick={handleResetIPv6} disabled={saving || !ipv6Unsaved}>
                    重置更改
                  </Button>
                  <Button variant="contained" color="primary" onClick={() => { void handleSaveIPv6(); }} disabled={saving || !ipv6Unsaved}>
                    保存
                  </Button>
                </Box>
              </Paper>
              
              {/* 多播设置 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4, border: multicastUnsaved ? '2px solid' : 'none', borderColor: 'warning.main' }}>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                  <Typography variant="h5">
                    多播设置
                  </Typography>
                  {multicastUnsaved && (
                    <Typography variant="body2" color="warning.main">
                      未保存
                    </Typography>
                  )}
                </Box>
                
                <Grid container spacing={3}>
                  <Grid size={{ xs: 12, md: 6 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                      <Typography variant="body1">
                        多播接收者限制
                      </Typography>
                      <TextField
                        type="number"
                        size="small"
                        value={multicastLimit}
                        onChange={(e) => setMulticastLimit(parseInt(e.target.value) || 0)}
                        sx={{ width: 120 }}
                      />
                    </Box>
                  </Grid>
                  <Grid size={{ xs: 12, md: 6 }}>
                    <FormControlLabel
                      control={<Switch checked={enableBroadcast} onChange={(e) => setEnableBroadcast(e.target.checked)} />}
                      label="启用广播"
                    />
                  </Grid>
                </Grid>
                
                <Box sx={{ display: 'flex', justifyContent: 'flex-start', gap: 2, mt: 3 }}>
                  <Button variant="outlined" onClick={handleResetMulticast} disabled={saving || !multicastUnsaved}>
                    重置更改
                  </Button>
                  <Button variant="contained" color="primary" onClick={() => { void handleSaveMulticast(); }} disabled={saving || !multicastUnsaved}>
                    保存
                  </Button>
                </Box>
              </Paper>
              
              {/* 删除网络 */}
              <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
                  <Typography variant="h5" sx={{ mb: 2, color: 'error.main' }}>
                    删除网络
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    此操作不可恢复。删除网络将断开所有连接的设备，并永久删除网络配置。
                  </Typography>
                  <Button variant="contained" color="error" size="large" onClick={() => setDeleteDialogOpen(true)}>
                    删除网络
                  </Button>
                </Paper>

                {/* 删除确认对话框 */}
                <Dialog
                  open={deleteDialogOpen}
                  onClose={() => setDeleteDialogOpen(false)}
                >
                  <DialogTitle>
                    确认删除网络
                  </DialogTitle>
                  <DialogContent>
                    <DialogContentText>
                      您确定要删除网络 "{network?.name}" 吗？此操作不可恢复，将永久删除该网络及其所有配置。
                    </DialogContentText>
                  </DialogContent>
                  <DialogActions>
                    <Button onClick={() => setDeleteDialogOpen(false)} color="primary">
                      取消
                    </Button>
                    <Button onClick={() => { void handleDeleteNetwork(); }} color="error" autoFocus disabled={saving}>
                      {saving ? '删除中...' : '确认删除'}
                    </Button>
                  </DialogActions>
                </Dialog>
            </>
          )}
        </>
      )}

      <Menu
        anchorEl={memberMenuAnchor}
        open={Boolean(memberMenuAnchor)}
        onClose={closeMemberMenu}
      >
        <MenuItem onClick={handleOpenEditMember}>编辑成员</MenuItem>
        {selectedMember && !selectedMember.authorized && (
          <MenuItem onClick={() => { void handleUpdateMemberStatus(selectedMember, true); }}>
            授权
          </MenuItem>
        )}
        {selectedMember && (
          <MenuItem onClick={() => { void handleUpdateMemberStatus(selectedMember, false); }}>
            拒绝设备
          </MenuItem>
        )}
        <MenuItem
          onClick={() => {
            setMemberDeleteDialogOpen(true);
            closeMemberMenu();
          }}
          sx={{ color: 'error.main' }}
        >
          移除成员
        </MenuItem>
      </Menu>

      <Dialog open={memberDialogOpen} onClose={() => setMemberDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>编辑网络成员</DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ mt: 1 }}>
            <Paper variant="outlined" sx={{ p: 2 }}>
              <Typography variant="h6" sx={{ mb: 2 }}>
                设备详情
              </Typography>
              <Stack spacing={2}>
                <Box sx={{ display: 'flex', gap: 1.5, alignItems: 'flex-start' }}>
                  <TextField
                    fullWidth
                    label="设备 ID"
                    value={selectedMember?.id || ''}
                    InputProps={{ readOnly: true }}
                  />
                  <Button variant="outlined" onClick={() => { void handleCopyMemberID(); }}>
                    复制
                  </Button>
                </Box>
                <TextField
                  fullWidth
                  label="设备名称"
                  value={memberForm.name}
                  onChange={(event) => setMemberForm((prev) => ({ ...prev, name: event.target.value }))}
                />
                <FormControlLabel
                  control={
                    <Switch
                      checked={memberForm.authorized}
                      onChange={(event) => setMemberForm((prev) => ({ ...prev, authorized: event.target.checked }))}
                    />
                  }
                  label={memberForm.authorized ? '设备已授权' : '设备待授权'}
                />
              </Stack>
            </Paper>

            <Paper variant="outlined" sx={{ p: 2 }}>
              <Typography variant="h6" sx={{ mb: 2 }}>
                成员元信息
              </Typography>
              <Grid container spacing={2}>
                <Grid size={{ xs: 12, sm: 6 }}>
                  <Typography variant="body2" color="text.secondary">ZeroTier 版本</Typography>
                  <Typography variant="body1">{selectedMember?.clientVersion || 'unknown'}</Typography>
                </Grid>
                <Grid size={{ xs: 12, sm: 6 }}>
                  <Typography variant="body2" color="text.secondary">最后活动</Typography>
                  <Typography variant="body1">{selectedMember?.lastSeenLabel || '未知'}</Typography>
                </Grid>
              </Grid>
            </Paper>

            <Paper variant="outlined" sx={{ p: 2 }}>
              <Typography variant="h6" sx={{ mb: 2 }}>
                Managed IPs
              </Typography>
              <Stack spacing={1.5}>
                {memberForm.ipAssignments.map((ip, index) => (
                  <Box key={`${selectedMember?.id || 'member'}-ip-${index}`} sx={{ display: 'flex', gap: 1.5 }}>
                    <TextField
                      fullWidth
                      label={`IP ${index + 1}`}
                      placeholder="例如 10.22.2.1"
                      value={ip}
                      onChange={(event) => handleMemberIpChange(index, event.target.value)}
                    />
                    <Button
                      variant="outlined"
                      color="error"
                      onClick={() => handleRemoveMemberIP(index)}
                      disabled={memberForm.ipAssignments.length === 1 && ip.trim() === ''}
                    >
                      删除
                    </Button>
                  </Box>
                ))}
                <Box>
                  <Button variant="outlined" onClick={handleAddMemberIP}>
                    添加 IP
                  </Button>
                </Box>
              </Stack>
            </Paper>

            <Paper variant="outlined" sx={{ p: 2 }}>
              <Typography variant="h6" sx={{ mb: 2 }}>
                Advanced Settings
              </Typography>
              <Stack spacing={1}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={memberForm.noAutoAssignIps}
                      onChange={(event) => setMemberForm((prev) => ({ ...prev, noAutoAssignIps: event.target.checked }))}
                    />
                  }
                  label="禁止自动分配 IP"
                />
                <FormControlLabel
                  control={
                    <Switch
                      checked={memberForm.activeBridge}
                      onChange={(event) => setMemberForm((prev) => ({ ...prev, activeBridge: event.target.checked }))}
                    />
                  }
                  label="允许以桥接设备身份接入"
                />
              </Stack>
            </Paper>
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setMemberDialogOpen(false)}>取消</Button>
          <Button onClick={() => { void handleSaveMember(); }} variant="contained" disabled={saving}>
            保存
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={memberDeleteDialogOpen} onClose={() => setMemberDeleteDialogOpen(false)}>
        <DialogTitle>确认移除成员</DialogTitle>
        <DialogContent>
          <DialogContentText>
            确定要将成员 "{selectedMember?.name || selectedMember?.id}" 从网络中移除吗？此操作会删除该成员在当前网络中的记录。
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setMemberDeleteDialogOpen(false)}>取消</Button>
          <Button onClick={() => { void handleDeleteMember(); }} color="error" disabled={saving}>
            移除
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={3000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert
          onClose={handleCloseSnackbar}
          severity={snackbar.severity}
          variant="filled"
          sx={{ width: '100%' }}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}

export default NetworkDetail;
