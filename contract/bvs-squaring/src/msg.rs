use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Addr;

#[cw_serde]
pub struct InstantiateMsg {
    pub aggregator: Addr,
    pub state_bank: Addr,
    pub bvs_driver: Addr,
}

#[cw_serde]
pub enum ExecuteMsg {
    CreateNewTask { input: Addr },
    RespondToTask { task_id: u64, result: i64 },
    Set { key: String, value: String },
    ExecuteBvsOffchain { task_id: String },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(i64)]
    GetTaskInput { task_id: u64 },

    #[returns(i64)]
    GetTaskResult { task_id: u64 },
}
