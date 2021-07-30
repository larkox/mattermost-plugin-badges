import React from 'react';

import {BadgeDetails, BadgeID} from '../../types/badges';
import Client from '../../client/api';

import {RHSState} from '../../types/general';
import {RHS_STATE_MY, RHS_STATE_OTHER} from '../../constants';
import BadgeImage from '../utils/badge_image';

import {markdown} from 'utils/markdown';

import RHSScrollbars from './rhs_scrollbars';
import UserRow from './user_row';

import './badge_details.scss';

type Props = {
    badgeID: BadgeID | null;
    currentUserID: string;
    actions: {
        setRHSView: (view: RHSState) => void;
        setRHSUser: (user: string | null) => void;
        getCustomEmojiByName: (names: string) => void;
    };
}

type State = {
    loading: boolean;
    badge?: BadgeDetails | null;
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

    componentDidUpdate(prevProps: Props, prevState: State) {
        if (this.state.badge !== prevState.badge && this.state.badge) {
            this.props.actions.getCustomEmojiByName(this.state.badge!.name);
        }

        if (this.props.badgeID === prevProps.badgeID) {
            return;
        }

        if (this.props.badgeID === null) {
            return;
        }

        const c = new Client();
        if (!this.state.loading) {
            this.setState({loading: true});
        }

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
            <div className='BadgeDetails'>
                <div><b>{'Badge Details'}</b></div>
                <div className='badge-info'>
                    <span className='badge-icon'>
                        <BadgeImage
                            badge={badge}
                            size={32}
                        />
                    </span>
                    <div className='badge-text'>
                        <div className='badge-name'>{badge.name}</div>
                        <div className='badge-description'>{markdown(badge.description)}</div>
                        <div className='badge-type'>{'Type: ' + badge.type_name}</div>
                        <div className='created-by'>{`Created by: ${badge.created_by_username}`}</div>
                    </div>
                </div>
                <div><b>{'Granted to:'}</b></div>
                <RHSScrollbars>{content}</RHSScrollbars>
            </div>
        );
    }
}

export default BadgeDetailsComponent;
