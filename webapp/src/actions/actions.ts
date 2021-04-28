import {AnyAction, Dispatch} from 'redux';

import {getCurrentChannel} from 'mattermost-redux/selectors/entities/channels';
import {getCurrentTeamId} from 'mattermost-redux/selectors/entities/teams';
import {GetStateFunc} from 'mattermost-redux/types/actions';
import {Client4} from 'mattermost-redux/client';
import {IntegrationTypes} from 'mattermost-redux/action_types';

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

export function setTriggerId(triggerId: string) {
    return {
        type: IntegrationTypes.RECEIVED_DIALOG_TRIGGER_ID,
        data: triggerId,
    };
}

export function openGrant(user?: string, badge?: string) {
    return (dispatch: Dispatch<AnyAction>, getState: GetStateFunc) => {
        let command = '/badges grant'
        if (user) {
            command += ` --user ${user}`
        }

        if (badge) {
            command += ` --badge ${badge}`
        }

        clientExecuteCommand(dispatch, getState, command)

        return {data: true}
    }
}

export function openCreateType() {
    return (dispatch: Dispatch<AnyAction>, getState: GetStateFunc) => {
        let command = '/badges create type'
        clientExecuteCommand(dispatch, getState, command)

        return {data: true}
    }
}

export function openCreateBadge() {
    return (dispatch: Dispatch<AnyAction>, getState: GetStateFunc) => {
        let command = '/badges create badge'
        clientExecuteCommand(dispatch, getState, command)

        return {data: true}
    }
}

export async function clientExecuteCommand(dispatch: Dispatch<AnyAction>, getState: GetStateFunc, command: string) {
    let currentChannel = getCurrentChannel(getState());
    const currentTeamId = getCurrentTeamId(getState());

    // Default to town square if there is no current channel (i.e., if Mattermost has not yet loaded)
    if (!currentChannel) {
        currentChannel = await Client4.getChannelByName(currentTeamId, 'town-square');
    }

    const args = {
        channel_id: currentChannel?.id,
        team_id: currentTeamId,
    };

    try {
        //@ts-ignore Typing in mattermost-redux is wrong
        const data = await Client4.executeCommand(command, args);
        dispatch(setTriggerId(data?.trigger_id));
    } catch (error) {
        console.error(error); //eslint-disable-line no-console
    }
}
