import React, { createContext, useContext, useEffect, useState, useRef } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { useAuth } from './AuthContext';
import { api } from '../../lib/api';
import { setupCheckinReminder, notifyRouletteOpened } from '../notifications';

// Recuerda el último grupo activo para no rebotar siempre al primero.
const ACTIVE_GROUP_KEY = 'dydi.activeGroup';

// Date utility functions
function dateISO(d: Date) {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

function localDateISO() {
  return dateISO(new Date());
}

interface Group {
  id: string;
  name: string;
  invite_code: string;
  created_at: string;
}

interface Member {
  user_id: string;
  display_name: string;
}

export interface Checkin {
  user_id: string;
  display_name: string;
  habit_id: string;
  habit_name: string;
  icon_key: string;
  color: string;
  status: 'done' | 'pending' | 'missed';
  scheduled_time?: string;
  note?: string;
  checked_on?: string;
}

interface Streak {
  habit_id: string;
  current: number;
}

interface Habit {
  id: string;
  name: string;
  description?: string;
  icon_key: string;
  color?: string;
}

interface Proposal {
  id: string;
  group_id: string;
  type: 'add_habit' | 'remove_habit' | 'kick_member' | 'delete_group';
  habit_id?: string;
  target_user_id?: string;
  status?: string;
  vote_count: number;
  member_count: number;
  created_at: string;
  user_voted: boolean;
}

interface Debt {
  id: string;
  roulette_entry_id?: string;
  group_id: string;
  debtor_id: string;
  punishment_text: string;
  punishment_emoji?: string;
  scope: 'personal' | 'collective';
  status?: 'pending' | 'completed' | 'expired' | 'forgiven';
  expires_at: string;
}

interface RouletteEntry {
  id: string;
  group_id: string;
  debtor_id: string;
  debtor_name?: string;
  suggestion_deadline: string;
  spun_at?: string;
}

interface Suggestion {
  id: string;
  entry_id: string;
  suggester_id: string;
  text: string;
  emoji?: string;
}

interface AppContextType {
  // Group state
  group: Group | null;
  members: Member[];
  onlineMembers: Set<string>;
  myGroups: { id: string; name: string }[];
  loadMyGroups: () => Promise<void>;
  loadGroup: (id: string) => Promise<void>;
  switchGroup: (id: string) => Promise<void>;
  autoLoad: () => Promise<boolean>;
  createGroup: (name: string) => Promise<Group>;
  joinGroup: (groupID: string, inviteCode: string) => Promise<void>;
  leaveGroup: () => Promise<void>;
  resetGroup: () => void;

  // Habits state
  todayCheckins: Checkin[];
  streaks: Record<string, number>;
  weekHistory: Record<string, Set<string>>;
  loadToday: (groupID: string) => Promise<void>;
  loadWeekHistory: (groupID: string) => Promise<void>;
  loadStreaks: (userID: string) => Promise<void>;
  checkin: (groupID: string, habitID: string, note?: string) => Promise<void>;

  // Proposals state
  catalog: Habit[];
  proposals: Proposal[];
  resolvedProposals: Proposal[];
  voted: Set<string>;
  loadCatalog: () => Promise<void>;
  loadProposals: (groupID: string) => Promise<void>;
  loadResolvedProposals: (groupID: string) => Promise<void>;
  propose: (
    groupID: string,
    type: string,
    habitID?: string | null,
    targetUserID?: string | null
  ) => Promise<Proposal>;
  vote: (proposalID: string, approved: boolean) => Promise<void>;

  // Penalties state
  debts: Debt[];
  resolvedDebts: Debt[];
  eligible: Member[];
  openEntries: RouletteEntry[];
  activeEntry: RouletteEntry | null;
  suggestions: Suggestion[];
  loadDebts: (groupID: string) => Promise<void>;
  loadResolvedDebts: (groupID: string) => Promise<void>;
  loadEligible: (groupID: string) => Promise<void>;
  loadOpenEntries: (groupID: string) => Promise<void>;
  enterEntry: (entry: RouletteEntry) => void;
  openRoulette: (groupID: string, debtorID: string) => Promise<RouletteEntry>;
  loadSuggestions: (entryID: string) => Promise<void>;
  submitSuggestion: (entryID: string, text: string, emoji?: string | null) => Promise<Suggestion>;
  spin: (entryID: string) => Promise<any>;
  completeDebt: (debtID: string) => Promise<Debt>;
  forgiveDebt: (debtID: string) => Promise<Debt>;
  clearEntry: () => void;

  // Shared state
  loading: boolean;
  wsConnected: boolean;
  loadAllData: () => Promise<void>;
}

const AppContext = createContext<AppContextType | undefined>(undefined);

export function AppProvider({ children }: { children: React.ReactNode }) {
  const { session, user } = useAuth();
  const [loading, setLoading] = useState(false);

  // Group State
  const [group, setGroup] = useState<Group | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [onlineMembers, setOnlineMembers] = useState<Set<string>>(new Set());
  const [myGroups, setMyGroups] = useState<{ id: string; name: string }[]>([]);

  // Habits State
  const [todayCheckins, setTodayCheckins] = useState<Checkin[]>([]);
  const [streaks, setStreaks] = useState<Record<string, number>>({});
  const [weekHistory, setWeekHistory] = useState<Record<string, Set<string>>>({});

  // Proposals State
  const [catalog, setCatalog] = useState<Habit[]>([]);
  const [proposals, setProposals] = useState<Proposal[]>([]);
  const [resolvedProposals, setResolvedProposals] = useState<Proposal[]>([]);
  const [voted, setVoted] = useState<Set<string>>(new Set());

  // Penalties State
  const [debts, setDebts] = useState<Debt[]>([]);
  const [resolvedDebts, setResolvedDebts] = useState<Debt[]>([]);
  const [eligible, setEligible] = useState<Member[]>([]);
  const [openEntries, setOpenEntries] = useState<RouletteEntry[]>([]);
  const [activeEntry, setActiveEntry] = useState<RouletteEntry | null>(null);
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);

  // WebSocket connection state
  const [wsConnected, setWsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const wsClosedRef = useRef(true);
  const wsAttemptsRef = useRef(0);

  // ── GROUP METHODS ────────────────────────────────────────────────────────
  async function loadMyGroups() {
    try {
      const data = await api('/api/groups');
      setMyGroups(data || []);
    } catch (err) {
      console.error('Error loading my groups:', err);
    }
  }

  async function loadGroup(id: string) {
    const data = await api(`/api/groups/${id}`);
    const { members: mems, ...groupData } = data;
    setGroup(groupData);
    setMembers(mems || []);
    AsyncStorage.setItem(ACTIVE_GROUP_KEY, groupData.id).catch(() => {});
  }

  // Cambia de grupo activo; el useEffect sobre group.id recarga datos y WS.
  async function switchGroup(id: string) {
    if (id === group?.id) return;
    await loadGroup(id);
  }

  async function autoLoad(): Promise<boolean> {
    if (group?.id) return true;

    // Fetch user groups
    const data = await api('/api/groups');
    const groupsList = data || [];
    setMyGroups(groupsList);

    if (groupsList.length === 0) {
      return false;
    }

    const remembered = await AsyncStorage.getItem(ACTIVE_GROUP_KEY).catch(() => null);
    const target = groupsList.find((g: any) => g.id === remembered) ?? groupsList[0];
    await loadGroup(target.id);
    return true;
  }

  async function createGroup(name: string): Promise<Group> {
    const data = await api('/api/groups', {
      method: 'POST',
      body: JSON.stringify({ name }),
    });
    setGroup(data);
    setMembers([]);
    setMyGroups(prev => [...prev.filter(g => g.id !== data.id), { id: data.id, name: data.name }]);
    AsyncStorage.setItem(ACTIVE_GROUP_KEY, data.id).catch(() => {});
    return data;
  }

  async function joinGroup(groupID: string, inviteCode: string) {
    await api(`/api/groups/${groupID}/join`, {
      method: 'POST',
      body: JSON.stringify({ invite_code: inviteCode }),
    });
    await loadGroup(groupID);
    // Refresca la lista completa; el nombre real viene del servidor.
    loadMyGroups().catch(() => {});
  }

  async function leaveGroup() {
    if (!group?.id) return;
    await api(`/api/groups/${group.id}/leave`, { method: 'DELETE' });
    resetGroup();
  }

  function resetGroup() {
    AsyncStorage.removeItem(ACTIVE_GROUP_KEY).catch(() => {});
    setGroup(null);
    setMembers([]);
    setMyGroups([]);
    setOnlineMembers(new Set());
    setTodayCheckins([]);
    setStreaks({});
    setWeekHistory({});
    setProposals([]);
    setVoted(new Set());
    setDebts([]);
    setEligible([]);
    setOpenEntries([]);
    setActiveEntry(null);
    setSuggestions([]);
  }

  // ── HABITS METHODS ───────────────────────────────────────────────────────
  async function loadToday(groupID: string) {
    const data = await api(`/api/habits/checkins/${groupID}/today?date=${localDateISO()}`);
    setTodayCheckins(data || []);
  }

  async function loadWeekHistory(groupID: string) {
    const to = new Date();
    const from = new Date();
    from.setDate(to.getDate() - 6);
    
    const list = await api(`/api/habits/history/${groupID}?from=${dateISO(from)}&to=${dateISO(to)}`);
    const map: Record<string, Set<string>> = {};
    for (const e of list || []) {
      const key = `${e.user_id}:${e.habit_id}`;
      if (!map[key]) {
        map[key] = new Set();
      }
      map[key].add(e.checked_on);
    }
    setWeekHistory(map);
  }

  async function loadStreaks(userID: string) {
    try {
      const list = await api(`/api/habits/streaks/${userID}`);
      const best = Array.isArray(list) 
        ? list.reduce((max, s) => Math.max(max, s.current ?? 0), 0) 
        : 0;
      setStreaks(prev => ({ ...prev, [userID]: best }));
    } catch (err) {
      console.error('Error loading streaks:', err);
    }
  }

  async function checkin(groupID: string, habitID: string, note = '') {
    const body: any = { group_id: groupID, habit_id: habitID, checked_on: localDateISO() };
    if (note) body.note = note;
    await api('/api/habits/checkins', {
      method: 'POST',
      body: JSON.stringify(body),
    });
    await loadToday(groupID);
  }

  // ── PROPOSALS METHODS ────────────────────────────────────────────────────
  async function loadCatalog() {
    const data = await api('/api/habits');
    setCatalog(data || []);
  }

  async function loadProposals(groupID: string) {
    const data = await api(`/api/groups/${groupID}/proposals`);
    const list = data || [];
    setProposals(list);
    // Sync voted set from server — survives app reloads
    setVoted(new Set(list.filter((p: Proposal) => p.user_voted).map((p: Proposal) => p.id)));
  }

  async function loadResolvedProposals(groupID: string) {
    const data = await api(`/api/groups/${groupID}/proposals?status=resolved`);
    setResolvedProposals(data || []);
  }

  async function propose(
    groupID: string,
    type: string,
    habitID: string | null = null,
    targetUserID: string | null = null
  ): Promise<Proposal> {
    const body: any = { type };
    if (habitID) body.habit_id = habitID;
    if (targetUserID) body.target_user_id = targetUserID;

    const p = await api(`/api/groups/${groupID}/proposals`, {
      method: 'POST',
      body: JSON.stringify(body),
    });
    setProposals(prev => [p, ...prev]);
    return p;
  }

  async function vote(proposalID: string, approved: boolean) {
    await api(`/api/proposals/${proposalID}/vote`, {
      method: 'POST',
      body: JSON.stringify({ approved }),
    });
    setVoted(prev => {
      const next = new Set(prev);
      next.add(proposalID);
      return next;
    });
    setProposals(prev => 
      prev.map(p => {
        if (p.id === proposalID && approved) {
          return { ...p, vote_count: (p.vote_count ?? 0) + 1 };
        }
        return p;
      })
    );
    // Re-fetch the authoritative list: a proposal that just reached quorum flips
    // to status != 'open' server-side and drops out, so it stops showing.
    if (group?.id) await loadProposals(group.id);
  }

  // ── PENALTIES METHODS ────────────────────────────────────────────────────
  async function loadDebts(groupID: string) {
    const data = await api(`/api/penalties/${groupID}/debts`);
    setDebts(data || []);
  }

  // Historial: cumplidas, perdonadas y expiradas (las activas viven en debts).
  async function loadResolvedDebts(groupID: string) {
    const data = await api(`/api/penalties/${groupID}/debts?status=resolved`);
    setResolvedDebts(data || []);
  }

  async function loadEligible(groupID: string) {
    const data = await api(`/api/penalties/${groupID}/eligible`);
    setEligible(data || []);
  }

  // Ruletas abiertas y sin girar: visibles para TODO el grupo, aunque la
  // elegibilidad de la semana ya haya expirado (re-abrirlas daría 409).
  async function loadOpenEntries(groupID: string) {
    const data = await api(`/api/penalties/${groupID}/roulette`);
    setOpenEntries(data || []);
  }

  // Entra a una ruleta ya abierta sin re-abrirla vía POST.
  function enterEntry(entry: RouletteEntry) {
    setActiveEntry(entry);
  }

  async function openRoulette(groupID: string, debtorID: string): Promise<RouletteEntry> {
    const data = await api('/api/penalties/roulette', {
      method: 'POST',
      body: JSON.stringify({ group_id: groupID, debtor_id: debtorID }),
    });
    setActiveEntry(data);
    setOpenEntries(prev => (prev.find(e => e.id === data.id) ? prev : [data, ...prev]));
    return data;
  }

  async function loadSuggestions(entryID: string) {
    const data = await api(`/api/penalties/roulette/${entryID}/suggestions`);
    setSuggestions(data || []);
  }

  async function submitSuggestion(entryID: string, text: string, emoji: string | null = null): Promise<Suggestion> {
    const s = await api(`/api/penalties/roulette/${entryID}/suggestions`, {
      method: 'POST',
      body: JSON.stringify({ text, ...(emoji ? { emoji } : {}) }),
    });
    setSuggestions(prev => [...prev, s]);
    return s;
  }

  async function spin(entryID: string): Promise<any> {
    const result = await api(`/api/penalties/roulette/${entryID}/spin`, {
      method: 'POST',
    });
    
    // Add spun_at timestamp to local activeEntry
    setActiveEntry(prev => prev ? { ...prev, spun_at: new Date().toISOString() } : null);
    setOpenEntries(prev => prev.filter(e => e.id !== entryID));

    const added = Array.isArray(result) ? result : [result];
    setDebts(prev => {
      const next = [...prev];
      for (const d of added) {
        if (!next.find(x => x.id === d.id)) {
          next.unshift(d);
        }
      }
      return next;
    });

    return result;
  }

  // El deudor marca su propia deuda como cumplida; sale de las activas.
  async function completeDebt(debtID: string): Promise<Debt> {
    const debt = await api(`/api/penalties/debts/${debtID}/complete`, { method: 'POST' });
    setDebts(prev => prev.filter(d => d.id !== debt.id));
    return debt;
  }

  // Cualquier miembro menos el deudor perdona una deuda: misericordia del squad.
  async function forgiveDebt(debtID: string): Promise<Debt> {
    const debt = await api(`/api/penalties/debts/${debtID}/forgive`, { method: 'POST' });
    setDebts(prev => prev.filter(d => d.id !== debt.id));
    return debt;
  }

  function clearEntry() {
    setActiveEntry(null);
    setSuggestions([]);
  }

  // ── GLOBAL LOAD DATA ─────────────────────────────────────────────────────
  async function loadAllData() {
    if (!group?.id) return;
    setLoading(true);
    try {
      await Promise.all([
        loadToday(group.id),
        loadWeekHistory(group.id),
        loadCatalog(),
        loadProposals(group.id),
        loadDebts(group.id),
        loadEligible(group.id),
        loadOpenEntries(group.id),
      ]);
      
      // Load streaks for all group members
      const memberIDs = [...new Set(members.map(m => m.user_id))];
      if (user?.id) memberIDs.push(user.id);
      await Promise.all(memberIDs.map(id => loadStreaks(id)));
    } catch (err) {
      console.error('Error loading all data:', err);
    } finally {
      setLoading(false);
    }
  }

  // Reload all data if group changes
  useEffect(() => {
    if (group?.id) {
      loadAllData();
      setupCheckinReminder();
    }
  }, [group?.id]);

  // ── WEBSOCKET PRESENCE AND UPDATES ──────────────────────────────────────
  useEffect(() => {
    const token = session?.access_token;
    const groupID = group?.id;

    if (!token || !groupID) {
      // Clean up socket if no token or no group
      disconnectWS();
      return;
    }

    const wsUrlBase = process.env.EXPO_PUBLIC_WS_URL || 'wss://dydi-25hj.onrender.com';
    const wsUrl = `${wsUrlBase}/ws/${groupID}?token=${token}`;

    wsClosedRef.current = false;
    wsAttemptsRef.current = 0;
    connectWS(wsUrl);

    return () => {
      disconnectWS();
    };

    function connectWS(url: string) {
      if (wsClosedRef.current) return;

      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        setWsConnected(true);
        wsAttemptsRef.current = 0;
        if (reconnectTimerRef.current) {
          clearTimeout(reconnectTimerRef.current);
          reconnectTimerRef.current = null;
        }
      };

      ws.onmessage = ({ data }) => {
        try {
          const msg = JSON.parse(data);
          
          switch (msg.type) {
            case 'checkin': {
              const checkinPayload = msg.payload as Checkin;
              setTodayCheckins(prev => {
                const idx = prev.findIndex(
                  c => c.user_id === checkinPayload.user_id && c.habit_id === checkinPayload.habit_id
                );
                if (idx >= 0) {
                  const next = [...prev];
                  next[idx] = { ...next[idx], ...checkinPayload };
                  return next;
                }
                return prev;
              });
              // El streak_update llega enseguida con la racha ya recalculada
              // por el backend — aquí ya no se re-consulta /streaks.
              break;
            }
            case 'streak_update': {
              const streakPayload = msg.payload;
              if (streakPayload.userID != null) {
                setStreaks(prev => ({ ...prev, [streakPayload.userID]: streakPayload.streak }));
              }
              break;
            }
            case 'member_online': {
              setOnlineMembers(prev => {
                const next = new Set(prev);
                next.add(msg.userID);
                return next;
              });
              break;
            }
            case 'member_offline': {
              setOnlineMembers(prev => {
                const next = new Set(prev);
                next.delete(msg.userID);
                return next;
              });
              break;
            }
            case 'roulette_start': {
              const entryPayload = msg.payload as RouletteEntry;
              setOpenEntries(prev =>
                prev.find(e => e.id === entryPayload.id) ? prev : [entryPayload, ...prev]
              );
              const isMe = entryPayload.debtor_id === session?.user?.id;
              const name =
                members.find(m => m.user_id === entryPayload.debtor_id)?.display_name ??
                'un miembro';
              notifyRouletteOpened(name, isMe);
              break;
            }
            case 'roulette_result':
            case 'collective_punishment': {
              const roulettePayload = msg.payload;
              setActiveEntry(prev => prev ? { ...prev, spun_at: new Date().toISOString() } : null);
              const addedDebts = Array.isArray(roulettePayload) ? roulettePayload : [roulettePayload];

              setDebts(prev => {
                const next = [...prev];
                for (const d of addedDebts) {
                  if (!next.find(x => x.id === d.id)) {
                    next.unshift(d);
                  }
                }
                return next;
              });
              setOpenEntries(prev =>
                prev.filter(e => !addedDebts.some((d: Debt) => d.roulette_entry_id === e.id))
              );
              break;
            }
            case 'debt_created': {
              const debtPayload = msg.payload as Debt;
              setDebts(prev => {
                if (!prev.find(d => d.id === debtPayload.id)) {
                  return [debtPayload, ...prev];
                }
                return prev;
              });
              break;
            }
            case 'debt_updated': {
              const debtPayload = msg.payload as Debt;
              // Las deudas activas solo listan status=pending: al resolverse, sale.
              setDebts(prev =>
                debtPayload.status !== 'pending'
                  ? prev.filter(d => d.id !== debtPayload.id)
                  : prev.map(d => (d.id === debtPayload.id ? debtPayload : d))
              );
              break;
            }
            default:
              break;
          }
        } catch (e) {
          console.error('Error parsing WebSocket message:', e);
        }
      };

      ws.onclose = () => {
        setWsConnected(false);
        if (!wsClosedRef.current) {
          scheduleReconnect(url);
        }
      };

      ws.onerror = () => {
        ws.close();
      };
    }

    function scheduleReconnect(url: string) {
      if (wsClosedRef.current || wsAttemptsRef.current >= 10) return;
      const wait = Math.min(1000 * Math.pow(2, wsAttemptsRef.current), 30_000);
      wsAttemptsRef.current += 1;
      
      if (reconnectTimerRef.current) clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = setTimeout(() => connectWS(url), wait);
    }

    function disconnectWS() {
      wsClosedRef.current = true;
      setWsConnected(false);
      
      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
      
      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.close();
        wsRef.current = null;
      }
    }
  }, [session?.access_token, group?.id]);

  return (
    <AppContext.Provider
      value={{
        group,
        members,
        onlineMembers,
        myGroups,
        loadMyGroups,
        loadGroup,
        switchGroup,
        autoLoad,
        createGroup,
        joinGroup,
        leaveGroup,
        resetGroup,

        todayCheckins,
        streaks,
        weekHistory,
        loadToday,
        loadWeekHistory,
        loadStreaks,
        checkin,

        catalog,
        proposals,
        resolvedProposals,
        voted,
        loadCatalog,
        loadProposals,
        loadResolvedProposals,
        propose,
        vote,

        debts,
        resolvedDebts,
        eligible,
        openEntries,
        activeEntry,
        suggestions,
        loadDebts,
        loadResolvedDebts,
        loadEligible,
        loadOpenEntries,
        enterEntry,
        openRoulette,
        loadSuggestions,
        submitSuggestion,
        spin,
        completeDebt,
        forgiveDebt,
        clearEntry,

        loading,
        wsConnected,
        loadAllData,
      }}
    >
      {children}
    </AppContext.Provider>
  );
}

export function useApp() {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error('useApp must be used within an AppProvider');
  }
  return context;
}
