import {UserProfile} from 'mattermost-redux/types/users';
import React from 'react';

import {Tooltip, OverlayTrigger} from 'react-bootstrap';

import {GlobalState} from 'mattermost-redux/types/store';

import {systemEmojis} from 'mattermost-redux/actions/emojis';

import {BadgeID, UserBadge} from 'types/badges';
import Client from 'client/api';
import BadgeImage from '../utils/badge_image';
import {RHSState} from 'types/general';
import {IMAGE_TYPE_EMOJI, RHS_STATE_DETAIL, RHS_STATE_MY, RHS_STATE_OTHER} from '../../constants';
import {markdown} from 'utils/markdown';

import './badge_list.scss';

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
        openGrant: (user?: string, badge?: string) => Promise<void>;
        getCustomEmojisByName: (names: string[]) => Promise<unknown>;
    };
}

type State = {
    badges?: UserBadge[];
    loaded?: boolean;
}

const MAX_BADGES = 7;
const BADGE_SIZE = 24;

class BadgeList extends React.PureComponent<Props, State> {
    constructor(props: Props) {
        super(props);

        this.state = {};
    }

    componentDidMount() {
        const c = new Client();
        c.getUserBadges(this.props.user.id).then((badges) => {
            this.setState({badges, loaded: true});
        });
    }

    componentDidUpdate(prevProps: Props, prevState: State) {
        if (this.state.badges !== prevState.badges) {
            const nBadges = this.state.badges?.length || 0;
            const toShow = nBadges < MAX_BADGES ? nBadges : MAX_BADGES;
            const names: string[] = [];
            for (let i = 0; i < toShow; i++) {
                const badge = this.state.badges![i];
                if (badge.image_type === IMAGE_TYPE_EMOJI) {
                    names.push(badge.image);
                }
            }
            const toLoad = names.filter((v) => !systemEmojis.has(v));
            this.props.actions.getCustomEmojisByName(toLoad);
        }
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
        this.props.hide();
    }

    onBadgeClick = (badge: UserBadge) => {
        if (!this.props.openRHS) {
            return;
        }

        this.props.actions.setRHSBadge(badge.id);
        this.props.actions.setRHSView(RHS_STATE_DETAIL);
        this.props.openRHS();
        this.props.hide();
    }

    onGrantClick = () => {
        this.props.actions.openGrant(this.props.user.username);
        this.props.hide();
    }

    render() {
        const nBadges = this.state.badges?.length || 0;
        const toShow = nBadges < MAX_BADGES ? nBadges : MAX_BADGES;

        const content: React.ReactNode[] = [];
        let row: React.ReactNode[] = [];
        for (let i = 0; i < toShow; i++) {
            const badge = this.state.badges![i];
            if (i !== 0) {
                content.push((<div>{row}</div>));
                row = [];
            }

            const time = new Date(badge.time);
            let reason = null;
            if (badge.reason) {
                reason = (<div>{'Why? ' + badge.reason}</div>);
            }
            const badgeComponent = (
                <OverlayTrigger
                    overlay={<Tooltip>
                        <div>{badge.name}</div>
                        <div>{markdown(badge.description)}</div>
                        {reason}
                        <div>{`Granted by: ${badge.granted_by_name}`}</div>
                        <div>{`Granted at: ${time.toDateString()}`}</div>
                    </Tooltip>}
                >
                    <span>
                        <a onClick={() => this.onBadgeClick(badge)}>
                            <BadgeImage
                                badge={badge}
                                size={BADGE_SIZE}
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
                <OverlayTrigger
                    overlay={<Tooltip>{`and ${nBadges - MAX_BADGES} more. Click to see all.`}</Tooltip>}
                >
                    <button
                        id='showMoreButton'
                        onClick={this.onMoreClick}
                    >
                        <span className={'fa fa-angle-right'}/>
                    </button>
                </OverlayTrigger>
            );
        }
        const maxWidth = (MAX_BADGES * BADGE_SIZE) + 30;
        let loading: React.ReactNode = null;
        if (!this.state.loaded) {
            loading = (

                // Reserve enough height one row of badges and the "and more" button
                <div style={{height: BADGE_SIZE, minWidth: 66, maxWidth}}>
                    {'Loading...'}
                </div>
            );
        }
        return (
            <div id='badgePlugin'>
                <div><b>{'Badges'}</b></div>
                <div id='contentContainer' >
                    {content}
                    {andMore}
                </div>
                {loading}
                <button
                    id='grantBadgeButton'
                    onClick={this.onGrantClick}
                >
                    <span className={'fa fa-plus-circle'}/>
                    {'Grant badge'}
                </button>
                <hr className='divider divider--expanded'/>
            </div>
        );
    }
}

export default BadgeList;
