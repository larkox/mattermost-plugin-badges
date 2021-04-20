import React from 'react';

import Scrollbars from 'react-custom-scrollbars';

import {BadgeDetails, BadgeID} from '../../types/badges';
import Client from '../../client/api';

import {RHSState} from '../../types/general';
import {RHS_STATE_MY, RHS_STATE_OTHER} from '../../constants';
import BadgeImage from '../utils/badge_image';

import UserRow from './user_row';

type Props = {
    badgeID: BadgeID | null,
    currentUserID: string,
    actions: {
        setRHSView: (view: RHSState) => void
        setRHSUser: (user: string | null) => void
    }
}

type State = {
    loading: boolean;
    badge?: BadgeDetails | null;
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

class BadgeDetailsComponent extends React.PureComponent<Props, State> {
    constructor(props: Props) {
        super(props);

        this.state = {
            loading: true,
        };
    }

    componentDidMount() {
        if (this.props.badgeID === null) {
            return;
        }
        
        const c = new Client();
        c.getBadgeDetails(this.props.badgeID).then((badge) => {
            this.setState({badge, loading: false});
        });
    }

    componentDidUpdate(prevProps: Props) {
        // Typical usage (don't forget to compare props):
        if (this.props.badgeID === prevProps.badgeID) {
            return;
        }

        if (this.props.badgeID === null) {
            this.setState({badge: null});
            return;
        }

        const c = new Client();
        c.getBadgeDetails(this.props.badgeID).then((badge) => {
            this.setState({badge, loading: false});
        });
    }

    onUserClick = (user: string) => {
        if (user === this.props.currentUserID) {
            this.props.actions.setRHSView(RHS_STATE_MY);
            return;
        }
        
        this.props.actions.setRHSUser(user);
        this.props.actions.setRHSView(RHS_STATE_OTHER);
    }

    render() {
        const {badge, loading} = this.state;
        if (this.props.badgeID == null) {
            return (<div>{'Badge not found.'}</div>);
        }

        if (loading) {
            return (<div>{'Loading...'}</div>);
        }

        if (!badge) {
            return (<div>{'Badge not found.'}</div>);
        }

        const row: React.ReactNode[] = [];
        const content = badge.owners.map((ownership) => {
            return (
                <UserRow
                    key={ownership.time}
                    ownership={ownership}
                    onClick={this.onUserClick}
                />
            );
        });
        return (
            <div style={{height: '100%'}}>
                <div><b>{'Badge Details'}</b></div>
                <div>
                    <span>
                        <BadgeImage
                            badge={badge}
                            size={32}
                        />
                    </span>
                    <div>{badge.name}</div>
                    <div>{badge.description}</div>
                    <div>{`Created by: ${badge.created_by_username}`}</div>
                </div>
                <div><b>{'Granted to:'}</b></div>
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

export default BadgeDetailsComponent;
