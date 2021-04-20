// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

import {connect} from 'react-redux';
import {ActionCreatorsMapObject, bindActionCreators, Dispatch} from 'redux';

import {GlobalState} from 'mattermost-redux/types/store';

import {getCurrentUserId} from 'mattermost-redux/selectors/entities/common';

import {setRHSView, setRHSBadge, setRHSUser} from '../../actions/actions';

import {getShowRHS} from 'selectors';
import {RHSState} from 'types/general';
import {BadgeID} from 'types/badges';

import BadgeList from './badge_list';

function mapStateToProps(state: GlobalState) {
    return {
        openRHS: getShowRHS(state),
        currentUserID: getCurrentUserId(state),
        debug: state,
    };
}

type Actions = {
    setRHSView: (view: RHSState) => Promise<void>;
    setRHSBadge: (id: BadgeID | null) => Promise<void>;
    setRHSUser: (id: string | null) => Promise<void>;
}

function mapDispatchToProps(dispatch: Dispatch) {
    return {
        actions: bindActionCreators<ActionCreatorsMapObject, Actions>({
            setRHSView,
            setRHSBadge,
            setRHSUser,
        }, dispatch),
    };
}

export default connect(mapStateToProps, mapDispatchToProps)(BadgeList);
