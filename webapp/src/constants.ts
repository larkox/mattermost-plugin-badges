import {PluginState, RHSState} from 'types/general';

export const IMAGE_TYPE_EMOJI = 'emoji';
export const IMAGE_TYPE_RELATIVE_URL = 'rel_url';
export const IMAGE_TYPE_ABSOLUTE_URL = 'abs_url';

export const RHS_STATE_MY: RHSState = 'my';
export const RHS_STATE_OTHER: RHSState = 'other';
export const RHS_STATE_ALL: RHSState = 'all';
export const RHS_STATE_DETAIL: RHSState = 'detail';

export const initialState: PluginState = {
    showRHS: null,
    rhsView: RHS_STATE_MY,
    rhsBadge: null,
    rhsUser: null,
};
