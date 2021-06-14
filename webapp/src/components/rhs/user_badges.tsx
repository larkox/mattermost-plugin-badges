import {UserProfile} from 'mattermost-redux/types/users';
import React from 'react';

import {BadgeID, UserBadge} from '../../types/badges';
import Client from '../../client/api';

import {RHSState} from 'types/general';
import {RHS_STATE_DETAIL} from '../../constants';

import UserBadgeRow from './user_badge_row';
import RHSScrollbars from './rhs_scrollbars';

import './user_badges.scss';

type Props = {
    isCurrentUser: boolean;
    user: UserProfile | null;
    actions: {
        setRHSView: (view: RHSState) => void;
        setRHSBadge: (badge: BadgeID | null) => void;
    };
}

type State = {
    loading: boolean;
    badges?: UserBadge[];
}

class UserBadges extends React.PureComponent<Props, State> {
    constructor(props: Props) {
        super(props);

        this.state = {
            loading: true,
        };
    }

    componentDidMount() {
        if (!this.props.user) {
            return;
        }
        const c = new Client();
        c.getUserBadges(this.props.user.id).then((badges) => {
            this.setState({badges, loading: false});
        });
    }

    componentDidUpdate(prevProps: Props) {
        if (this.props.user === prevProps.user) {
            return;
        }

        if (!this.props.user) {
            return;
        }

        const c = new Client();
        if (!this.state.loading) {
            this.setState({loading: true});
        }

        c.getUserBadges(this.props.user.id).then((badges) => {
            this.setState({badges, loading: false});
        });
    }

    onBadgeClick = (badge: UserBadge) => {
        this.props.actions.setRHSBadge(badge.id);
        this.props.actions.setRHSView(RHS_STATE_DETAIL);
    }

    render() {
        if (!this.props.user) {
            return (<div>{'User not found.'}</div>);
        }

        if (this.state.loading) {
            return (<div>{'Loading...'}</div>);
        }

        if (!this.state.badges || this.state.badges.length === 0) {
            return (<div>{'No badges yet.'}</div>);
        }

        const content = this.state.badges.map((badge) => {
            return (
                <UserBadgeRow
                    isCurrentUser={this.props.isCurrentUser}
                    key={badge.time}
                    badge={badge}
                    onClick={this.onBadgeClick}
                />
            );
        });

        let title = 'My badges';
        if (!this.props.isCurrentUser) {
            title = `@${this.props.user.username}'s badges`;
        }
        return (
            <div className='UserBadges'>
                <div><b>{title}</b></div>
                <RHSScrollbars>{content}</RHSScrollbars>
            </div>
        );
    }
}

export default UserBadges;
