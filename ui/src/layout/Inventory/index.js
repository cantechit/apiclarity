import React, { useState } from 'react';
import { Route, Switch, useRouteMatch, Redirect } from 'react-router-dom';
import MainTitleWithRefresh from 'components/MainTitleWithRefresh';
import TabbedPageContainer from 'components/TabbedPageContainer';
import InventoryTable from './InventoryTable';
import InventoryDetails from './InventoryDetails';
import GeneralFilter, { formatFiltersToQueryParams } from './GeneralFilter';

import './inventory.scss';

export const API_TYPE_ITEMS = {
    INTERNAL: {value: "INTERNAL", label: "Internal"},
    EXTERNAL: {value: "EXTERNAL", label: "External"}
};

const getInternalPath = path => `${path}/${API_TYPE_ITEMS.INTERNAL.value}`;
const getExternalPath = path => `${path}/${API_TYPE_ITEMS.EXTERNAL.value}`;

const Inventory = () => {
    const {path} = useRouteMatch();

    const [filters, setFilters] = useState([]);
    const paramsFilters = formatFiltersToQueryParams(filters);

    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = () => setRefreshTimestamp(Date());

    return (
        <div className="inventory-tables-page">
            <MainTitleWithRefresh title="API Inventory" onRefreshClick={doRefreshTimestamp} />
            <GeneralFilter
                filters={filters}
                onFilterUpdate={filters => setFilters(filters)}
            />
            <TabbedPageContainer
                items={[
                    {
                        title: "Internal",
                        to: path,
                        exact: true,
                        component: () => <InventoryTable basePath={path} type={API_TYPE_ITEMS.INTERNAL.value} filters={paramsFilters} refreshTimestamp={refreshTimestamp} />
                    },
                    {
                        title: "External",
                        to: getExternalPath(path),
                        component: () => <InventoryTable basePath={path} type={API_TYPE_ITEMS.EXTERNAL.value} filters={paramsFilters} refreshTimestamp={refreshTimestamp} />
                    }
                ]}
            />
        </div>
    )
}

const InventoryRouter = () => {
    const {path} = useRouteMatch();
    const internalPath = getInternalPath(path);
    const externalPath = getExternalPath(path);

    return (
        <Switch>
            <Redirect exact from={internalPath} to={path} />
            <Route path={`${internalPath}/:inventoryId`} component={() => <InventoryDetails type={API_TYPE_ITEMS.INTERNAL.value} />} />
            <Route path={`${externalPath}/:inventoryId`} component={() => <InventoryDetails type={API_TYPE_ITEMS.EXTERNAL.value} />} />
            <Route path={path} component={Inventory} />
        </Switch>
    )
}

export default InventoryRouter;