import {combineReducers} from 'redux';

import {GenericAction} from 'mattermost-redux/types/actions';

import ActionTypes from '../action_types';
import {RHS_STATE_MY} from '../constants';

function showRHS(state = null, action: GenericAction) {
    switch (action.type) {
    case ActionTypes.RECEIVED_SHOW_RHS_ACTION:
        return action.showRHSPluginAction;
    default:
        return state;
    }
}

function rhsView(state = RHS_STATE_MY, action: GenericAction) {
    switch (action.type) {
    case ActionTypes.RECEIVED_RHS_VIEW:
        return action.data;
    default:
        return state;
    }
}

function rhsUser(state = null, action: GenericAction) {
    switch (action.type) {
    case ActionTypes.RECEIVED_RHS_USER:
        return action.data;
    default:
        return state;
    }
}

function rhsBadge(state = null, action: GenericAction) {
    switch (action.type) {
    case ActionTypes.RECEIVED_RHS_BADGE:
        return action.data;
    default:
        return state;
    }
}

export default combineReducers({
    showRHS,
    rhsView,
    rhsUser,
    rhsBadge,
});
