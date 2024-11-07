use crate::{
    error::ContractError,
    msg::{ExecuteMsg, InstantiateMsg, QueryMsg},
    state::{
        AGGREGATOR, BVS_DRIVER, CREATED_TASKS, MAX_ID, OPERATOR_MAX_SCORE, OPERATOR_SCORE,
        RESPONDED_TASKS, STATE_BANK,
    },
};

use cosmwasm_std::{
    entry_point, to_json_binary, Addr, Binary, CosmosMsg, Deps, DepsMut, Env, Event, MessageInfo,
    Response, WasmMsg,
};
use cw2::set_contract_version;

const CONTRACT_NAME: &str = env!("CARGO_PKG_NAME");
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
    MAX_ID.save(deps.storage, &0)?;
    AGGREGATOR.save(deps.storage, &msg.aggregator)?;
    STATE_BANK.save(deps.storage, &msg.state_bank)?;
    BVS_DRIVER.save(deps.storage, &msg.bvs_driver)?;

    let response = Response::new().add_attribute("method", "instantiate");
    Ok(response)
}

#[entry_point]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::CreateNewTask { input } => create_new_task(deps, info, input),
        ExecuteMsg::RespondToTask { task_id, result } => {
            respond_to_task(deps, info, task_id, result)
        }
        _ => Ok(Response::default()),
    }
}

fn create_new_task(
    deps: DepsMut,
    _info: MessageInfo,
    input: Addr,
) -> Result<Response, ContractError> {
    let id = MAX_ID.may_load(deps.storage)?;
    let new_id = id.unwrap_or(0) + 1;

    MAX_ID.save(deps.storage, &new_id)?;

    CREATED_TASKS.save(deps.storage, new_id, &input)?;

    // construct message for calling method in the state bank contract
    let state_bank_address = STATE_BANK.load(deps.storage)?;
    let mut msg = ExecuteMsg::Set {
        key: format!("taskId.{}", new_id),
        value: input.to_string(),
    };

    let mut wasm_msg = WasmMsg::Execute {
        contract_addr: state_bank_address.into_string(),
        msg: to_json_binary(&msg)?,
        funds: vec![],
    };

    let state_bank_msg = CosmosMsg::Wasm(wasm_msg);

    // construct message for calling method in the bvs driver contract
    let bvs_driver_address = BVS_DRIVER.load(deps.storage)?;
    msg = ExecuteMsg::ExecuteBvsOffchain {
        task_id: new_id.to_string(),
    };

    wasm_msg = WasmMsg::Execute {
        contract_addr: bvs_driver_address.into_string(),
        msg: to_json_binary(&msg)?,
        funds: vec![],
    };

    let bvs_driver_msg = CosmosMsg::Wasm(wasm_msg);

    // emit event
    let event = Event::new("NewTaskCreated")
        .add_attribute("taskId", new_id.to_string())
        .add_attribute("input", input.to_string());

    Ok(Response::new()
        .add_message(state_bank_msg)
        .add_message(bvs_driver_msg)
        .add_attribute("method", "CreateNewTask")
        .add_attribute("input", input.to_string())
        .add_attribute("taskId", new_id.to_string())
        .add_event(event))
}

fn respond_to_task(
    deps: DepsMut,
    info: MessageInfo,
    task_id: u64,
    result: i64,
) -> Result<Response, ContractError> {
    let aggregator = AGGREGATOR.load(deps.storage)?;

    // only aggregator can respond to task
    if info.sender != aggregator {
        return Err(ContractError::Unauthorized {});
    }

    let responded_result = RESPONDED_TASKS.may_load(deps.storage, task_id)?;

    // result already submitted
    if let Some(_) = responded_result {
        return Err(ContractError::ResultSubmitted {});
    }

    // save task result
    RESPONDED_TASKS.save(deps.storage, task_id, &result)?;

    // fetch operator address for the task
    let operator = CREATED_TASKS.load(deps.storage, task_id)?;

    // update the operator's score based on the result
    OPERATOR_SCORE.update(
        deps.storage,
        operator.clone(),
        |score| -> Result<_, ContractError> {
            let current_score = score.unwrap_or(0); // Default score is 0 if not set
            let updated_score = if result == 1 {
                current_score + 1
            } else {
                current_score - 1
            };
            Ok(updated_score)
        },
    )?;

    // increment operator's max score
    OPERATOR_MAX_SCORE.update(
        deps.storage,
        operator.clone(),
        |max_score| -> Result<_, ContractError> { Ok(max_score.unwrap_or(0) + 1) },
    )?;

    // emit event
    let event = Event::new("TaskResponded")
        .add_attribute("taskId", task_id.to_string())
        .add_attribute("result", result.to_string());

    Ok(Response::new()
        .add_attribute("method", "RespondToTask")
        .add_attribute("taskId", task_id.to_string())
        .add_attribute("result", result.to_string())
        .add_event(event))
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> Result<Binary, ContractError> {
    match msg {
        QueryMsg::GetTaskInput { task_id } => query_task_input(deps, task_id),
        QueryMsg::GetTaskResult { task_id } => query_task_result(deps, task_id),
        QueryMsg::GetOperatorScore { operator } => query_operator_score(deps, operator),
    }
}

fn query_task_input(deps: Deps, task_id: u64) -> Result<Binary, ContractError> {
    let result = CREATED_TASKS.may_load(deps.storage, task_id)?;

    if let Some(input) = result {
        return Ok(to_json_binary(&input)?);
    }

    Err(ContractError::NoValueFound {})
}

fn query_task_result(deps: Deps, task_id: u64) -> Result<Binary, ContractError> {
    let result = RESPONDED_TASKS.may_load(deps.storage, task_id)?;

    if let Some(result) = result {
        return Ok(to_json_binary(&result)?);
    }

    Err(ContractError::NoValueFound {})
}

fn query_operator_score(deps: Deps, operator: Addr) -> Result<Binary, ContractError> {
    let score = OPERATOR_SCORE.may_load(deps.storage, operator)?;

    if let Some(score) = score {
        return Ok(to_json_binary(&score)?);
    }

    Err(ContractError::NoValueFound {})
}
#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::{
        from_json,
        testing::{mock_dependencies, mock_env},
        Addr, Coin, CosmosMsg, MessageInfo, WasmMsg,
    };

    fn mock_info(sender: &str, funds: &[Coin]) -> MessageInfo {
        MessageInfo {
            sender: Addr::unchecked(sender),
            funds: funds.to_vec(),
        }
    }

    #[test]
    fn test_instantiate() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);
        let msg = InstantiateMsg {
            aggregator: Addr::unchecked("aggregator"),
            state_bank: Addr::unchecked("state_bank"),
            bvs_driver: Addr::unchecked("bvs_driver"),
        };

        let res = instantiate(deps.as_mut(), env, info, msg).unwrap();
        assert_eq!(0, res.messages.len());

        // Check if the contract version is set
        let version = cw2::get_contract_version(deps.as_ref().storage).unwrap();
        assert_eq!(version.contract, CONTRACT_NAME);
        assert_eq!(version.version, CONTRACT_VERSION);

        // Check if the state is properly initialized
        assert_eq!(MAX_ID.load(deps.as_ref().storage).unwrap(), 0);
        assert_eq!(
            AGGREGATOR.load(deps.as_ref().storage).unwrap(),
            Addr::unchecked("aggregator")
        );
        assert_eq!(
            STATE_BANK.load(deps.as_ref().storage).unwrap(),
            Addr::unchecked("state_bank")
        );
        assert_eq!(
            BVS_DRIVER.load(deps.as_ref().storage).unwrap(),
            Addr::unchecked("bvs_driver")
        );
    }

    #[test]
    fn test_create_new_task() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);
        let msg = InstantiateMsg {
            aggregator: Addr::unchecked("aggregator"),
            state_bank: Addr::unchecked("state_bank"),
            bvs_driver: Addr::unchecked("bvs_driver"),
        };
        instantiate(deps.as_mut(), env.clone(), info.clone(), msg).unwrap();
        let create_msg = ExecuteMsg::CreateNewTask {
            input: Addr::unchecked("operator"),
        };
        let res = execute(deps.as_mut(), env, info, create_msg).unwrap();

        // Check if the task was created successfully
        assert_eq!(2, res.messages.len());
        assert_eq!(1, res.events.len());

        // Check if MAX_ID was incremented
        assert_eq!(MAX_ID.load(deps.as_ref().storage).unwrap(), 1);

        // Check if the task was saved
        assert_eq!(
            CREATED_TASKS.load(deps.as_ref().storage, 1).unwrap(),
            Addr::unchecked("operator")
        );

        // Check if the correct messages were created
        let state_bank_msg = &res.messages[0];
        let bvs_driver_msg = &res.messages[1];

        match &state_bank_msg.msg {
            CosmosMsg::Wasm(WasmMsg::Execute {
                contract_addr, msg, ..
            }) => {
                assert_eq!(contract_addr, "state_bank");
                let parsed_msg: ExecuteMsg = from_json(msg).unwrap();
                match parsed_msg {
                    ExecuteMsg::Set { key, value } => {
                        assert_eq!(key, "taskId.1");
                        assert_eq!(value, Addr::unchecked("operator").to_string());
                    }
                    _ => panic!("Unexpected message type"),
                }
            }
            _ => panic!("Unexpected message type"),
        }

        match &bvs_driver_msg.msg {
            CosmosMsg::Wasm(WasmMsg::Execute {
                contract_addr, msg, ..
            }) => {
                assert_eq!(contract_addr, "bvs_driver");
                let parsed_msg: ExecuteMsg = from_json(msg).unwrap();
                match parsed_msg {
                    ExecuteMsg::ExecuteBvsOffchain { task_id } => {
                        assert_eq!(task_id, "1");
                    }
                    _ => panic!("Unexpected message type"),
                }
            }
            _ => panic!("Unexpected message type"),
        }
    }

    #[test]
    fn test_respond_to_task() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);
        let msg = InstantiateMsg {
            aggregator: Addr::unchecked("aggregator"),
            state_bank: Addr::unchecked("state_bank"),
            bvs_driver: Addr::unchecked("bvs_driver"),
        };
        instantiate(deps.as_mut(), env.clone(), info.clone(), msg).unwrap();

        // Create the first task
        let create_msg_1 = ExecuteMsg::CreateNewTask {
            input: Addr::unchecked("operator"),
        };
        execute(deps.as_mut(), env.clone(), info.clone(), create_msg_1).unwrap();

        // Respond to the first task with a positive result (score should increase)
        let respond_msg_1 = ExecuteMsg::RespondToTask {
            task_id: 1,
            result: 1,
        };
        let aggregator_info = mock_info("aggregator", &[]);
        let res = execute(
            deps.as_mut(),
            env.clone(),
            aggregator_info.clone(),
            respond_msg_1,
        )
        .unwrap();

        // Check if the first response was saved successfully
        assert_eq!(0, res.messages.len());
        assert_eq!(1, res.events.len());
        assert_eq!(RESPONDED_TASKS.load(deps.as_ref().storage, 1).unwrap(), 1);

        // Verify that the operator's score has been incremented
        let operator_addr = Addr::unchecked("operator");
        assert_eq!(
            OPERATOR_SCORE
                .load(deps.as_ref().storage, operator_addr.clone())
                .unwrap(),
            1
        );

        // Verify that OPERATOR_MAX_SCORE was incremented
        assert_eq!(
            OPERATOR_MAX_SCORE
                .load(deps.as_ref().storage, operator_addr.clone())
                .unwrap(),
            1
        );

        // Create the second task
        let create_msg_2 = ExecuteMsg::CreateNewTask {
            input: Addr::unchecked("operator"),
        };
        execute(deps.as_mut(), env.clone(), info, create_msg_2).unwrap();

        // Respond to the second task with a negative result (score should decrease)
        let respond_msg_2 = ExecuteMsg::RespondToTask {
            task_id: 2,
            result: 0,
        };
        let res = execute(deps.as_mut(), env, aggregator_info, respond_msg_2).unwrap();

        // Check if the second response was saved successfully
        assert_eq!(0, res.messages.len());
        assert_eq!(1, res.events.len());
        assert_eq!(RESPONDED_TASKS.load(deps.as_ref().storage, 2).unwrap(), 0);

        // Verify that the operator's score has been decremented
        assert_eq!(
            OPERATOR_SCORE
                .load(deps.as_ref().storage, operator_addr.clone())
                .unwrap(),
            0
        );

        // Verify that OPERATOR_MAX_SCORE was incremented again
        assert_eq!(
            OPERATOR_MAX_SCORE
                .load(deps.as_ref().storage, operator_addr.clone())
                .unwrap(),
            2
        );
    }

    #[test]
    fn respond_to_task_unauthorized() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);
        let msg = InstantiateMsg {
            aggregator: Addr::unchecked("aggregator"),
            state_bank: Addr::unchecked("state_bank"),
            bvs_driver: Addr::unchecked("bvs_driver"),
        };
        instantiate(deps.as_mut(), env.clone(), info.clone(), msg).unwrap();

        // Create a new task
        let create_msg = ExecuteMsg::CreateNewTask {
            input: Addr::unchecked("operator"),
        };
        execute(deps.as_mut(), env.clone(), info, create_msg).unwrap();

        // Try to respond to the task with an unauthorized address
        let respond_msg = ExecuteMsg::RespondToTask {
            task_id: 1,
            result: 84,
        };
        let unauthorized_info = mock_info("unauthorized", &[]);
        let res = execute(deps.as_mut(), env, unauthorized_info, respond_msg);

        // Check if the response was rejected
        assert!(res.is_err());
        match res.unwrap_err() {
            ContractError::Unauthorized {} => {}
            e => panic!("Unexpected error: {:?}", e),
        }
    }

    #[test]
    fn query_task_input() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);
        let msg = InstantiateMsg {
            aggregator: Addr::unchecked("aggregator"),
            state_bank: Addr::unchecked("state_bank"),
            bvs_driver: Addr::unchecked("bvs_driver"),
        };
        instantiate(deps.as_mut(), env.clone(), info.clone(), msg).unwrap();

        // Create a new task
        let create_msg = ExecuteMsg::CreateNewTask {
            input: Addr::unchecked("operator"),
        };
        execute(deps.as_mut(), env.clone(), info, create_msg).unwrap();

        // Query the task input
        let query_msg = QueryMsg::GetTaskInput { task_id: 1 };
        let res = query(deps.as_ref(), env, query_msg).unwrap();
        let value: String = from_json(&res).unwrap();
        assert_eq!(value, Addr::unchecked("operator").to_string());
    }

    #[test]
    fn query_task_result() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);
        let msg = InstantiateMsg {
            aggregator: Addr::unchecked("aggregator"),
            state_bank: Addr::unchecked("state_bank"),
            bvs_driver: Addr::unchecked("bvs_driver"),
        };
        instantiate(deps.as_mut(), env.clone(), info.clone(), msg).unwrap();

        // Create a new task
        let create_msg = ExecuteMsg::CreateNewTask {
            input: Addr::unchecked("operator"),
        };
        execute(deps.as_mut(), env.clone(), info, create_msg).unwrap();

        // Respond to the task
        let respond_msg = ExecuteMsg::RespondToTask {
            task_id: 1,
            result: 84,
        };
        let aggregator_info = mock_info("aggregator", &[]);
        execute(deps.as_mut(), env.clone(), aggregator_info, respond_msg).unwrap();

        // Query the task result
        let query_msg = QueryMsg::GetTaskResult { task_id: 1 };
        let res = query(deps.as_ref(), env, query_msg).unwrap();
        let value: i64 = from_json(&res).unwrap();
        assert_eq!(value, 84);
    }

    #[test]
    fn query_operator_score() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);
        let msg = InstantiateMsg {
            aggregator: Addr::unchecked("aggregator"),
            state_bank: Addr::unchecked("state_bank"),
            bvs_driver: Addr::unchecked("bvs_driver"),
        };
        instantiate(deps.as_mut(), env.clone(), info.clone(), msg).unwrap();

        // Create a new task and respond to it to increase the operator's score
        let create_msg = ExecuteMsg::CreateNewTask {
            input: Addr::unchecked("operator"),
        };
        execute(deps.as_mut(), env.clone(), info.clone(), create_msg).unwrap();

        // Respond to the task to update the operator's score
        let respond_msg = ExecuteMsg::RespondToTask {
            task_id: 1,
            result: 1,
        };
        let aggregator_info = mock_info("aggregator", &[]);
        execute(deps.as_mut(), env.clone(), aggregator_info, respond_msg).unwrap();

        // Query the operator's score
        let operator_addr = Addr::unchecked("operator");
        let query_msg = QueryMsg::GetOperatorScore {
            operator: operator_addr,
        };
        let res = query(deps.as_ref(), env, query_msg).unwrap();

        // Deserialize the result into the score (i64)
        let value: i64 = from_json(&res).unwrap();
        assert_eq!(value, 1);
    }
}
/*
babylond tx wasm execute bbn14esj8vqlpugc04k3l30n7td70ywrm5370pp74m64uhr5yujtrs7qnq09ay '{"create_new_task": {"input": 10}}' --from=bvs-operator --gas=auto --gas-prices=1ubbn --gas-adjustment=1.3 --chain-id=sat-bbn-testnet1 -b=sync --yes --log_format=json --node https://rpc.sat-bbn-testnet1.satlayer.net
babylond tx wasm execute bbn1h9zjs2zr2xvnpngm9ck8ja7lz2qdt5mcw55ud7wkteycvn7aa4pqpghx2q '{"add_registered_bvs_contract":{"address":"bbn1ssjvqd8z7dh93djet606079c6ge7t922lkzyvjgrtjvgrnhgh9cs3lyu6f"}}' --from=bvs-owner --gas=auto --gas-prices=1ubbn --gas-adjustment=1.3 --chain-id=sat-bbn-testnet1 -b=sync --yes --log_format=json --node https://rpc.sat-bbn-testnet1.satlayer.net
babylond tx wasm instantiate 74 '{"aggregator": "bbn14awtreyhrhnq97dzsyngepax8gu9qtxysap3wz", "state_bank": "bbn1h9zjs2zr2xvnpngm9ck8ja7lz2qdt5mcw55ud7wkteycvn7aa4pqpghx2q", "bvs_driver": "bbn18x5lx5dda7896u074329fjk4sflpr65s036gva65m4phavsvs3rqk5e59c"}' --from=bvs-owner --no-admin --label="demo" --gas=auto --gas-prices=1ubbn --gas-adjustment=1.3 --chain-id=sat-bbn-testnet1 -b=sync --yes --log_format=json --node https://rpc.sat-bbn-testnet1.satlayer.net
curl -s -X GET "https://lcd1.sat-bbn-testnet1.satlayer.net/cosmwasm/wasm/v1/code/74/contracts" -H "accept: application/json" | jq -r '.contracts[-1]'
 */
