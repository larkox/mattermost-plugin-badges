import ActionTypes from 'action_types/';
import {BadgeID} from 'types/badges';
import {RHSState} from 'types/general';

/**
 * Stores`showRHSPlugin` action returned by
 * registerRightHandSidebarComponent in plugin initialization.
 */
export function setShowRHSAction(showRHSPluginAction: () => void) {
    return {
        type: ActionTypes.RECEIVED_SHOW_RHS_ACTION,
        showRHSPluginAction,
    };
}

export function setRHSUser(userID: string | null) {
    return {
        type: ActionTypes.RECEIVED_RHS_USER,
        data: userID,
    };
}

export function setRHSBadge(badgeID: BadgeID | null) {
    return {
        type: ActionTypes.RECEIVED_RHS_BADGE,
        data: badgeID,
    };
}

export function setRHSView(view: RHSState) {
    return {
        type: ActionTypes.RECEIVED_RHS_VIEW,
        data: view,
    };
}
