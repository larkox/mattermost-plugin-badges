import {createSelector} from 'reselect';

import {getCustomEmojisByName} from 'mattermost-redux/selectors/entities/emojis';

import EmojiMap from '../utils/emoji_map';
import {id as pluginId} from '../manifest';
import {initialState} from '../constants';
import {PluginState} from 'types/general';

const getPluginState = (state: any): PluginState => state['plugins-' + pluginId] || initialState;

export const getEmojiMap = createSelector(
    getCustomEmojisByName,
    (customEmojisByName) => {
        return new EmojiMap(customEmojisByName);
    },
);

export const getShowRHS = createSelector(
    getPluginState,
    (state) => {
        return state.showRHS;
    },
);

export const getRHSView = createSelector(
    getPluginState,
    (state) => {
        return state.rhsView;
    },
);

export const getRHSUser = createSelector(
    getPluginState,
    (state) => {
        return state.rhsUser;
    },
);

export const getRHSBadge = createSelector(
    getPluginState,
    (state) => {
        return state.rhsBadge;
    },
);
