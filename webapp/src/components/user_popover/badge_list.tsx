import {UserProfile} from 'mattermost-redux/types/users';
import React from 'react';

import {Tooltip, OverlayTrigger} from 'react-bootstrap';

import {GlobalState} from 'mattermost-redux/types/store';

import {BadgeID, UserBadge} from 'types/badges';
import Client from 'client/api';
import BadgeImage from '../utils/badge_image';
import {RHSState} from 'types/general';
import {RHS_STATE_DETAIL, RHS_STATE_MY, RHS_STATE_OTHER} from '../../constants';
import {markdown} from 'utils/markdown';

type Props = {
    debug: GlobalState;
    user: UserProfile;
    currentUserID: string;
    openRHS: (() => void) | null;
    hide: () => void;
    status?: string;
    actions: {
        setRHSView: (view: RHSState) => Promise<void>;
        setRHSBadge: (id: BadgeID | null) => Promise<void>;
        setRHSUser: (id: string | null) => Promise<void>;
    };
}

type State = {
    badges?: UserBadge[];
}

const MAX_BADGES = 20;
const BADGES_PER_ROW = 10;

class BadgeList extends React.PureComponent<Props, State> {
    constructor(props: Props) {
        super(props);

        this.state = {};
    }

    componentDidMount() {
        const c = new Client();
        c.getUserBadges(this.props.user.id).then((badges) => {
            this.setState({badges});
        });
    }

    onMoreClick = () => {
        if (!this.props.openRHS) {
            return;
        }

        if (this.props.currentUserID === this.props.user.id) {
            this.props.actions.setRHSView(RHS_STATE_MY);
            this.props.openRHS();
            return;
        }

        this.props.actions.setRHSUser(this.props.user.id);
        this.props.actions.setRHSView(RHS_STATE_OTHER);
        this.props.openRHS();
    }

    onBadgeClick = (badge: UserBadge) => {
        if (!this.props.openRHS) {
            return;
        }

        this.props.actions.setRHSBadge(badge.id);
        this.props.actions.setRHSView(RHS_STATE_DETAIL);
        this.props.openRHS();
    }

    render() {
        if (!this.state.badges || this.state.badges.length === 0) {
            return null;
        }

        const nBadges = this.state.badges?.length || 0;
        const toShow = nBadges < MAX_BADGES ? nBadges : MAX_BADGES;

        const content: React.ReactNode[] = [];
        let row: React.ReactNode[] = [];
        for (let i = 0; i < toShow; i++) {
            const badge = this.state.badges![i];
            if (i !== 0 && i % BADGES_PER_ROW === 0) {
                content.push((<div>{row}</div>));
                row = [];
            }

            const time = new Date(badge.time);
            const badgeComponent = (
                <OverlayTrigger
                    overlay={<Tooltip>
                        <div>{badge.name}</div>
                        <div>{markdown(badge.description)}</div>
                        <div>{`Granted by: ${badge.granted_by_name}`}</div>
                        <div>{`Granted at: ${time.toDateString()}`}</div>
                    </Tooltip>}
                >
                    <span>
                        <a onClick={() => this.onBadgeClick(badge)}>
                            <BadgeImage
                                badge={badge}
                                size={24}
                            />
                        </a>
                    </span>
                </OverlayTrigger>
            );
            row.push(badgeComponent);
        }
        content.push((<div>{row}</div>));
        let andMore: React.ReactNode = null;
        if (nBadges > MAX_BADGES) {
            andMore = (
                <a onClick={this.onMoreClick}>
                    <div>{`and ${nBadges - MAX_BADGES} more`}</div>
                </a>
            );
        }
        return (
            <div>
                <div><b>{'Badges'}</b></div>
                {content}
                {andMore}
            </div>
        );
    }
}

export default BadgeList;
