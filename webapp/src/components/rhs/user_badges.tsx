import {UserProfile} from 'mattermost-redux/types/users';
import React from 'react';

import Scrollbars from 'react-custom-scrollbars';

import {BadgeID, UserBadge} from '../../types/badges';
import Client from '../../client/api';

import {RHSState} from 'types/general';
import {RHS_STATE_DETAIL} from '../../constants';

import UserBadgeRow from './user_badge_row';

type Props = {
    title: string,
    user: UserProfile | null,
    actions: {
        setRHSView: (view: RHSState) => void
        setRHSBadge: (badge: BadgeID | null) => void
    }
}

type State = {
    loading: boolean;
    badges?: UserBadge[];
}

function renderView(props: any) {
    return (
        <div
            {...props}
            className='scrollbar--view'
        />);
}

function renderThumbHorizontal(props: any) {
    return (
        <div
            {...props}
            className='scrollbar--horizontal'
        />);
}

function renderThumbVertical(props: any) {
    return (
        <div
            {...props}
            className='scrollbar--vertical'
        />);
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
        // Typical usage (don't forget to compare props):
        if (this.props.user === prevProps.user) {
            return;
        }

        if (!this.props.user) {
            this.setState({badges: []});
            return;
        }

        const c = new Client();
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

        const row: React.ReactNode[] = [];
        const content = this.state.badges.map((badge) => {
            return (
                <UserBadgeRow
                    key={badge.time}
                    badge={badge}
                    onClick={this.onBadgeClick}
                />
            );
        });
        return (
            <div style={{height: '100%'}}>
                <div><b>{this.props.title}</b></div>
                <Scrollbars
                    autoHide={true}
                    autoHideTimeout={500}
                    autoHideDuration={500}
                    renderThumbHorizontal={renderThumbHorizontal}
                    renderThumbVertical={renderThumbVertical}
                    renderView={renderView}
                >
                    {content}
                </Scrollbars>
            </div>
        );
    }
}

export default UserBadges;
