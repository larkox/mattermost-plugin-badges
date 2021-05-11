import {Store} from 'redux';

import {GlobalState} from 'mattermost-redux/types/store';

import {GenericAction} from 'mattermost-redux/types/actions';

import React from 'react';

import {openAddSubscription, openCreateBadge, openCreateType, openRemoveSubscription, setRHSView, setShowRHSAction} from 'actions/actions';

import UserBadges from 'components/rhs';

import ChannelHeaderButton from 'components/channel_header_button';

import Reducer from './reducers';

import manifest from './manifest';

// eslint-disable-next-line import/no-unresolved
import {PluginRegistry} from './types/mattermost-webapp';
import BadgeList from './components/user_popover/';
import {RHS_STATE_ALL} from './constants';

export default class Plugin {
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, GenericAction>) {
        registry.registerReducer(Reducer);

        registry.registerPopoverUserAttributesComponent(BadgeList);

        const {showRHSPlugin, toggleRHSPlugin} = registry.registerRightHandSidebarComponent(UserBadges, 'Badges');
        store.dispatch(setShowRHSAction(() => store.dispatch(showRHSPlugin)));

        registry.registerChannelHeaderButtonAction(
            <ChannelHeaderButton/>,
            () => {
                store.dispatch(setRHSView(RHS_STATE_ALL));
                store.dispatch(toggleRHSPlugin);
            },
            'Badges',
            'Open your list of badges.',
        );

        registry.registerMainMenuAction(
            'Create badge',
            () => {
                store.dispatch(openCreateBadge() as any);
            },
            null,
        );
        registry.registerMainMenuAction(
            'Create badge type',
            () => {
                store.dispatch(openCreateType() as any);
            },
            null,
        );

        registry.registerChannelHeaderMenuAction(
            'Add badge subscription',
            () => {
                store.dispatch(openAddSubscription() as any);
            },
        );
        registry.registerChannelHeaderMenuAction(
            'Remove badge subscription',
            () => {
                store.dispatch(openRemoveSubscription() as any);
            },
        );
    }
}

declare global {
    interface Window {
        registerPlugin(id: string, plugin: Plugin): void;
    }
}

window.registerPlugin(manifest.id, new Plugin());
