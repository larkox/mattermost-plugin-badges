// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import {useDispatch, useSelector} from 'react-redux';

import {getCurrentUser} from 'mattermost-redux/selectors/entities/common';

import React from 'react';

import {getUser} from 'mattermost-redux/selectors/entities/users';

import {GlobalState} from 'mattermost-redux/types/store';

import {getRHSBadge, getRHSUser, getRHSView} from 'selectors';
import {RHS_STATE_ALL, RHS_STATE_DETAIL, RHS_STATE_OTHER, RHS_STATE_MY} from '../../constants';
import {RHSState} from 'types/general';
import {setRHSBadge, setRHSUser, setRHSView} from 'actions/actions';
import {BadgeID} from 'types/badges';

import UserBadges from './user_badges';
import BadgeDetailsComponent from './badge_details';
import AllBadges from './all_badges';

const RHS: React.FC = () => {
    const dispatch = useDispatch();
    const currentView = useSelector(getRHSView);
    const currentBadge = useSelector(getRHSBadge);
    const currentUserID = useSelector(getRHSUser);
    const currentUser = useSelector((state: GlobalState) => getUser(state, (currentUserID as string)));
    const myUser = useSelector(getCurrentUser);

    switch (currentView) {
    case RHS_STATE_ALL:
        return (
            <AllBadges
                actions={{
                    setRHSView: (view: RHSState) => dispatch(setRHSView(view)),
                    setRHSBadge: (badge: BadgeID | null) => dispatch(setRHSBadge(badge)),
                }}
            />
        );
    case RHS_STATE_DETAIL:
        return (
            <BadgeDetailsComponent
                badgeID={currentBadge}
                currentUserID={myUser.id}
                actions={{
                    setRHSView: (view: RHSState) => dispatch(setRHSView(view)),
                    setRHSUser: (user: string | null) => dispatch(setRHSUser(user)),
                }}
            />
        );
    case RHS_STATE_OTHER:
        return (
            <UserBadges
                user={currentUser}
                isCurrentUser={false}
                actions={{
                    setRHSView: (view: RHSState) => dispatch(setRHSView(view)),
                    setRHSBadge: (badge: BadgeID | null) => dispatch(setRHSBadge(badge)),
                }}
            />
        );
    case RHS_STATE_MY:
    default:
        return (
            <UserBadges
                user={myUser}
                isCurrentUser={true}
                actions={{
                    setRHSView: (view: RHSState) => dispatch(setRHSView(view)),
                    setRHSBadge: (badge: BadgeID | null) => dispatch(setRHSBadge(badge)),
                }}
            />
        );
    }
};

export default RHS;
