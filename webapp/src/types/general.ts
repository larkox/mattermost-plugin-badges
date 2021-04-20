import {BadgeID} from './badges';

export type RHSState = string;

export type PluginState = {
    showRHS: (() => void)| null;
    rhsView: RHSState;
    rhsUser: string | null;
    rhsBadge: BadgeID | null;
}
