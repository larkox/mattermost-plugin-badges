import React from 'react';

import {BadgeID, AllBadgesBadge} from '../../types/badges';
import Client from '../../client/api';

import {RHSState} from '../../types/general';
import {RHS_STATE_DETAIL} from '../../constants';

import AllBadgesRow from './all_badges_row';
import RHSScrollbars from './rhs_scrollbars';

import './all_badges.scss';

type Props = {
    actions: {
        setRHSView: (view: RHSState) => void;
        setRHSBadge: (badge: BadgeID | null) => void;
    };
}

type State = {
    loading: boolean;
    badges?: AllBadgesBadge[];
}

class AllBadges extends React.PureComponent<Props, State> {
    constructor(props: Props) {
        super(props);

        this.state = {
            loading: true,
        };
    }

    componentDidMount() {
        const c = new Client();
        c.getAllBadges().then((badges) => {
            this.setState({badges, loading: false});
        });
    }

    onBadgeClick = (badge: AllBadgesBadge) => {
        this.props.actions.setRHSBadge(badge.id);
        this.props.actions.setRHSView(RHS_STATE_DETAIL);
    }

    render() {
        if (this.state.loading) {
            return (<div className='AllBadges'>{'Loading...'}</div>);
        }

        if (!this.state.badges || this.state.badges.length === 0) {
            return (<div className='AllBadges'>{'No badges yet.'}</div>);
        }

        const content = this.state.badges.map((badge) => {
            return (
                <AllBadgesRow
                    key={badge.id}
                    badge={badge}
                    onClick={this.onBadgeClick}
                />
            );
        });
        return (
            <div className='AllBadges'>
                <div><b>{'All badges'}</b></div>
                <RHSScrollbars>{content}</RHSScrollbars>
            </div>
        );
    }
}

export default AllBadges;
