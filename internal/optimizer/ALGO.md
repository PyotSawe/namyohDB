# Query Optimizer Module Algorithms

## Overview
This document describes the core algorithms used in the SQL query optimizer for cost-based optimization, including logical plan transformation, physical plan generation, cardinality estimation, and cost modeling techniques.

## Core Optimization Algorithms

### 1. Two-Phase Optimization Algorithm

#### Algorithm: `optimizeQuery(ast, schema, statistics)`
```
Input: Abstract syntax tree, database schema, table statistics
Output: Optimized physical execution plan

Algorithm optimizeQuery(ast, schema, statistics):
  1. Logical optimization phase:
     a. Convert AST to logical plan
     b. Apply logical rewrite rules:
        - Predicate pushdown
        - Join reordering
        - Subquery elimination
        - Constant folding
     c. Simplify expressions
     d. Generate optimized logical plan
  
  2. Physical optimization phase:
     a. Generate physical alternatives for each logical operator
     b. Enumerate join algorithms (NLJ, Hash Join, SMJ)
     c. Select access paths (table scan vs. index scan)
     d. Apply cost model to each alternative
     e. Select minimum cost plan using dynamic programming
  
  3. Post-optimization:
     a. Validate plan correctness
     b. Cache plan for reuse
     c. Return optimized physical plan
```

**Time Complexity**: O(n³) for join ordering, O(k) for rule application where n is number of relations, k is plan complexity
**Space Complexity**: O(2ⁿ) for dynamic programming table in worst case

### 2. Logical Plan Transformation Algorithms

#### Algorithm: `applyPredicatePushdown(plan)`
```
Input: Logical plan tree
Output: Plan with predicates pushed down

Algorithm applyPredicatePushdown(plan):
  1. Traverse plan tree bottom-up
  2. For each Selection operator:
     a. Analyze predicate conditions
     b. Identify pushable predicates:
        - Simple column predicates
        - Predicates involving only one relation
        - Non-correlated predicates
     c. Push predicates down past:
        - Projections (if columns preserved)
        - Joins (to appropriate side based on column references)
        - Union operations (duplicate to both sides)
  3. Update plan tree structure
  4. Remove redundant Selection operators
```

#### Algorithm: `eliminateSubqueries(plan)`
```
Input: Logical plan with subqueries
Output: Plan with subqueries converted to joins

Algorithm eliminateSubqueries(plan):
  1. Identify subqueries in plan:
     a. Scalar subqueries in SELECT list
     b. EXISTS/NOT EXISTS subqueries
     c. IN/NOT IN subqueries
     d. Correlated subqueries
  
  2. Transform each subquery type:
     a. Scalar subquery:
        - Convert to LEFT OUTER JOIN
        - Add aggregation if needed
     
     b. EXISTS subquery:
        - Convert to SEMI JOIN
        - Ensure proper null handling
     
     c. IN subquery:
        - Convert to INNER JOIN if values are unique
        - Convert to SEMI JOIN otherwise
     
     d. Correlated subquery:
        - Identify correlation predicates
        - Convert to LATERAL JOIN or decorrelate using aggregation
  
  3. Update plan tree with new join structure
```

### 3. Join Reordering Algorithm

#### Algorithm: `optimizeJoinOrder(joinTree, statistics)`
```
Input: Join tree with n relations, table statistics
Output: Optimal join order

Algorithm optimizeJoinOrder(joinTree, statistics):
  1. Extract join predicates and relations
  2. If n <= 4: // Small joins - use exhaustive search
     a. Enumerate all possible join orders
     b. Cost each permutation
     c. Return minimum cost order
  
  3. If n > 4: // Large joins - use dynamic programming
     a. Initialize DP table dp[subset]
     b. Base case: dp[{R}] = cost(scan(R)) for each relation R
     
     c. For subset_size = 2 to n:
        For each subset S of size subset_size:
          dp[S] = min over all splits (S1, S2) of S:
            dp[S1] + dp[S2] + cost(join(S1, S2))
     
     d. Return plan corresponding to dp[all_relations]
  
  4. Apply join commutativity and associativity rules
  5. Consider available indexes for join conditions
```

**Join Ordering Complexity**:
- **Exhaustive**: O(n!) time, suitable for n ≤ 6
- **Dynamic Programming**: O(3ⁿ) time, O(2ⁿ) space, suitable for n ≤ 20
- **Heuristic Methods**: O(n²) time for larger joins

#### Algorithm: `selectJoinAlgorithm(leftRel, rightRel, joinCondition, statistics)`
```
Input: Two relations to join, join condition, statistics
Output: Best join algorithm

Algorithm selectJoinAlgorithm(leftRel, rightRel, joinCondition, statistics):
  1. Analyze join condition:
     a. Equi-join vs. non-equi join
     b. Number of join keys
     c. Selectivity of join
  
  2. Estimate relation sizes and costs:
     leftSize = estimateCardinality(leftRel)
     rightSize = estimateCardinality(rightRel)
     joinSelectivity = estimateJoinSelectivity(joinCondition)
  
  3. Calculate costs for each algorithm:
     a. Nested Loop Join:
        cost_NLJ = leftSize + leftSize * rightSize
     
     b. Hash Join:
        cost_Hash = leftSize + rightSize + hash_overhead
        
     c. Sort-Merge Join:
        cost_SMJ = leftSize * log(leftSize) + rightSize * log(rightSize) + 
                   leftSize + rightSize
  
  4. Consider available indexes:
     a. If index exists on join key:
        - Consider Index Nested Loop Join
        - Adjust costs based on index selectivity
  
  5. Return algorithm with minimum estimated cost
```

### 4. Cost-Based Plan Selection Algorithm

#### Algorithm: `enumeratePhysicalPlans(logicalPlan)`
```
Input: Logical plan node
Output: Set of equivalent physical plans with costs

Algorithm enumeratePhysicalPlans(logicalPlan):
  1. Generate physical alternatives based on logical operator:
     
     a. LogicalTableScan:
        - PhysicalTableScan (sequential scan)
        - PhysicalIndexScan (for each applicable index)
     
     b. LogicalJoin:
        - PhysicalNestedLoopJoin
        - PhysicalHashJoin (if equi-join)
        - PhysicalSortMergeJoin
     
     c. LogicalAggregation:
        - PhysicalHashAggregation
        - PhysicalSortAggregation
     
     d. LogicalSort:
        - PhysicalSort (if not already sorted)
        - PhysicalIndexScan (if index provides required order)
  
  2. For each alternative:
     a. Recursively generate child plans
     b. Calculate plan cost using cost model
     c. Prune dominated plans (higher cost, same properties)
  
  3. Return Pareto-optimal set of plans
```

## Cardinality Estimation Algorithms

### 5. Histogram-Based Selectivity Estimation

#### Algorithm: `estimateSelectivity(predicate, histogram)`
```
Input: Selection predicate, column histogram
Output: Estimated selectivity (0.0 to 1.0)

Algorithm estimateSelectivity(predicate, histogram):
  1. Parse predicate type:
     
     a. Equality predicate (col = value):
        If value in histogram:
          return frequency(value) / total_rows
        Else:
          return 1.0 / ndv(col)  // Uniform distribution assumption
     
     b. Range predicate (col > value):
        buckets_above = 0
        For each bucket in histogram:
          If bucket.upper_bound > value:
            If bucket.lower_bound > value:
              buckets_above += bucket.count
            Else:
              // Interpolate within bucket
              range_fraction = (bucket.upper_bound - value) / 
                              (bucket.upper_bound - bucket.lower_bound)
              buckets_above += bucket.count * range_fraction
        
        return buckets_above / total_rows
     
     c. LIKE predicate (col LIKE pattern):
        return estimateLikeSelectivity(pattern)
  
  2. For compound predicates:
     a. AND: multiply selectivities (independence assumption)
     b. OR: use inclusion-exclusion principle
```

#### Algorithm: `estimateJoinCardinality(leftCard, rightCard, joinCondition)`
```
Input: Left relation cardinality, right relation cardinality, join condition
Output: Estimated join result cardinality

Algorithm estimateJoinCardinality(leftCard, rightCard, joinCondition):
  1. Analyze join type:
     
     a. Cross product (no join condition):
        return leftCard * rightCard
     
     b. Equi-join on foreign key:
        return max(leftCard, rightCard)  // Each left tuple matches ≤1 right tuple
     
     c. General equi-join:
        ndv_left = getDistinctValues(left_join_column)
        ndv_right = getDistinctValues(right_join_column)
        ndv_join = min(ndv_left, ndv_right)
        
        return (leftCard * rightCard) / ndv_join
     
     d. Non-equi join:
        selectivity = estimatePredicateSelectivity(joinCondition)
        return leftCard * rightCard * selectivity
  
  2. Apply correlation adjustments if available
  3. Ensure result doesn't exceed maximum possible cardinality
```

### 6. Multi-Column Statistics Algorithm

#### Algorithm: `estimateMultiColumnSelectivity(predicates, correlationStats)`
```
Input: Multiple predicates on different columns, correlation statistics
Output: Combined selectivity estimate

Algorithm estimateMultiColumnSelectivity(predicates, correlationStats):
  1. If no correlation statistics available:
     // Use independence assumption
     combinedSelectivity = 1.0
     For each predicate p in predicates:
       combinedSelectivity *= estimateSelectivity(p)
     return combinedSelectivity
  
  2. If correlation statistics available:
     a. Group predicates by correlation groups
     b. For each correlation group:
        - Use joint distribution statistics
        - Calculate group selectivity directly
     c. For independent groups:
        - Multiply group selectivities
  
  3. Apply backoff strategy if correlation data incomplete:
     return alpha * correlated_estimate + (1-alpha) * independent_estimate
```

## Cost Model Algorithms

### 7. I/O Cost Estimation Algorithm

#### Algorithm: `estimateIOCost(operator, statistics)`
```
Input: Physical operator, database statistics
Output: Estimated I/O cost in page reads

Algorithm estimateIOCost(operator, statistics):
  1. Switch on operator type:
     
     a. TableScan:
        pages_to_scan = table.page_count
        selectivity = estimateSelectivity(filter_condition)
        if selectivity < 0.1:  // Selective scan
          return pages_to_scan  // Must read all pages
        else:
          return pages_to_scan  // Sequential scan cost
     
     b. IndexScan:
        index_levels = log_fanout(index.row_count)
        index_pages_read = index_levels  // Tree traversal
        
        selectivity = estimateIndexSelectivity(scan_condition)
        data_pages_read = selectivity * table.page_count
        
        return index_pages_read + data_pages_read
     
     c. NestedLoopJoin:
        outer_cost = estimateIOCost(outer_child)
        outer_card = estimateCardinality(outer_child)
        inner_cost_per_tuple = estimateIOCost(inner_child)
        
        return outer_cost + outer_card * inner_cost_per_tuple
     
     d. HashJoin:
        build_cost = estimateIOCost(build_child)
        probe_cost = estimateIOCost(probe_child)
        
        return build_cost + probe_cost
```

#### Algorithm: `estimateCPUCost(operator, cardinality)`
```
Input: Physical operator, estimated cardinality
Output: Estimated CPU cost in instruction cycles

Algorithm estimateCPUCost(operator, cardinality):
  1. Define base CPU costs per tuple:
     SCAN_CPU_COST = 1.0
     HASH_CPU_COST = 2.0
     COMPARISON_CPU_COST = 0.5
     AGGREGATION_CPU_COST = 3.0
  
  2. Calculate operator-specific CPU cost:
     
     a. TableScan with filter:
        return cardinality * (SCAN_CPU_COST + 
                             filter_conditions * COMPARISON_CPU_COST)
     
     b. HashJoin:
        build_cost = build_cardinality * HASH_CPU_COST
        probe_cost = probe_cardinality * HASH_CPU_COST
        return build_cost + probe_cost
     
     c. Sort:
        return cardinality * log2(cardinality) * COMPARISON_CPU_COST
     
     d. Aggregation:
        return cardinality * AGGREGATION_CPU_COST
```

### 8. Memory Cost Estimation Algorithm

#### Algorithm: `estimateMemoryCost(operator)`
```
Input: Physical operator
Output: Estimated memory requirement in bytes

Algorithm estimateMemoryCost(operator):
  1. Switch on operator type:
     
     a. HashJoin:
        build_table_size = build_cardinality * avg_tuple_size
        hash_overhead = build_table_size * 0.5  // Hash table overhead
        return build_table_size + hash_overhead
     
     b. Sort:
        input_size = input_cardinality * avg_tuple_size
        if input_size <= memory_limit:
          return input_size  // In-memory sort
        else:
          return memory_limit  // External sort with spilling
     
     c. HashAggregation:
        distinct_groups = estimateDistinctGroups(group_by_columns)
        group_state_size = distinct_groups * (group_key_size + aggregate_state_size)
        return group_state_size
```

## Advanced Optimization Algorithms

### 9. Dynamic Programming for Join Optimization

#### Algorithm: `dpJoinOptimization(relations, predicates)`
```
Input: Set of relations, join predicates
Output: Optimal join tree

Algorithm dpJoinOptimization(relations, predicates):
  1. Initialize DP table:
     dp = map[set_of_relations] -> BestPlan
     
  2. Base case - single relations:
     For each relation R in relations:
       dp[{R}] = createScanPlan(R)
  
  3. Build up subsets of increasing size:
     For subset_size = 2 to |relations|:
       For each subset S of size subset_size:
         dp[S] = null
         min_cost = infinity
         
         For each split (S1, S2) where S1 ∪ S2 = S, S1 ∩ S2 = ∅:
           If joinable(S1, S2, predicates):
             left_plan = dp[S1]
             right_plan = dp[S2]
             
             join_methods = [NestedLoop, Hash, SortMerge]
             For each method in join_methods:
               join_plan = createJoinPlan(method, left_plan, right_plan)
               cost = estimateCost(join_plan)
               
               If cost < min_cost:
                 min_cost = cost
                 dp[S] = join_plan
  
  4. Return dp[all_relations]
```

**Optimization**: Use branch-and-bound pruning to reduce search space

### 10. Genetic Algorithm for Large Join Optimization

#### Algorithm: `geneticJoinOptimization(relations, predicates, generations)`
```
Input: Relations, predicates, number of generations
Output: Near-optimal join plan

Algorithm geneticJoinOptimization(relations, predicates, generations):
  1. Initialize population:
     population = []
     For i = 1 to population_size:
       random_plan = generateRandomJoinTree(relations, predicates)
       population.append(random_plan)
  
  2. Evolve for specified generations:
     For generation = 1 to generations:
       a. Evaluate fitness (inverse of cost):
          For each plan in population:
            plan.fitness = 1.0 / estimateCost(plan)
       
       b. Selection (tournament selection):
          new_population = []
          For i = 1 to population_size:
            parent1 = tournamentSelect(population)
            parent2 = tournamentSelect(population)
            child = crossover(parent1, parent2)
            child = mutate(child, mutation_rate)
            new_population.append(child)
       
       c. population = new_population
  
  3. Return best plan from final population
```

### 11. Adaptive Statistics Update Algorithm

#### Algorithm: `adaptiveStatsUpdate(query_feedback)`
```
Input: Query execution feedback
Output: Updated statistics

Algorithm adaptiveStatsUpdate(query_feedback):
  1. Analyze estimation errors:
     For each table T in query_feedback.tables:
       estimated_card = feedback.estimated_cardinality[T]
       actual_card = feedback.actual_cardinality[T]
       error_ratio = abs(estimated_card - actual_card) / actual_card
       
       If error_ratio > error_threshold:
         trigger_stats_update(T)
  
  2. Update selectivity estimates:
     For each predicate P with feedback:
       old_selectivity = current_selectivity[P]
       actual_selectivity = feedback.actual_selectivity[P]
       
       // Exponential smoothing
       alpha = 0.1
       new_selectivity = alpha * actual_selectivity + (1-alpha) * old_selectivity
       current_selectivity[P] = new_selectivity
  
  3. Update correlation statistics:
     For each column pair (C1, C2) with joint predicate feedback:
       update_correlation_coefficient(C1, C2, feedback)
```

## Plan Caching Algorithms

### 12. Plan Cache Management Algorithm

#### Algorithm: `planCacheLookup(query_hash, current_stats)`
```
Input: Query hash, current database statistics
Output: Cached plan or null

Algorithm planCacheLookup(query_hash, current_stats):
  1. Check if plan exists in cache:
     cached_entry = cache.get(query_hash)
     If cached_entry == null:
       return null
  
  2. Validate plan freshness:
     For each table T in cached_entry.referenced_tables:
       cached_stats = cached_entry.statistics[T]
       current_table_stats = current_stats[T]
       
       // Check for significant changes
       row_count_change = abs(current_table_stats.row_count - cached_stats.row_count)
       If row_count_change / cached_stats.row_count > staleness_threshold:
         cache.invalidate(query_hash)
         return null
  
  3. Update access information:
     cached_entry.access_count += 1
     cached_entry.last_accessed = current_time()
     cache.updateLRU(query_hash)
  
  4. Return cached_entry.physical_plan
```

#### Algorithm: `planCacheEviction()`
```
Input: None (triggered when cache is full)
Output: Evicted plan entries

Algorithm planCacheEviction():
  1. Identify eviction candidates:
     candidates = []
     For each entry in cache:
       score = calculateEvictionScore(entry)
       candidates.append((entry, score))
  
  2. Sort by eviction score (higher score = more likely to evict):
     sort(candidates, key=lambda x: x.score, reverse=True)
  
  3. Evict entries until sufficient space:
     evicted = []
     freed_memory = 0
     For (entry, score) in candidates:
       If freed_memory >= required_memory:
         break
       cache.remove(entry.query_hash)
       evicted.append(entry)
       freed_memory += entry.memory_size
  
  4. Return evicted

calculateEvictionScore(entry):
  recency_score = 1.0 / (current_time() - entry.last_accessed)
  frequency_score = entry.access_count
  size_penalty = entry.memory_size / average_entry_size
  
  return size_penalty / (recency_score * frequency_score)
```

## Complexity Analysis

### Algorithm Complexity Summary

| Algorithm | Time Complexity | Space Complexity | Notes |
|-----------|-----------------|------------------|--------|
| Two-Phase Optimization | O(n³) | O(2ⁿ) | n = number of relations |
| Predicate Pushdown | O(p) | O(1) | p = number of predicates |
| Subquery Elimination | O(s × n) | O(n) | s = subqueries, n = plan nodes |
| Join Reordering (DP) | O(3ⁿ) | O(2ⁿ) | n = number of relations |
| Join Reordering (Genetic) | O(g × p × n) | O(p) | g = generations, p = population |
| Cardinality Estimation | O(h) | O(1) | h = histogram buckets |
| Cost Estimation | O(1) | O(1) | Per operator |
| Plan Cache Lookup | O(1) avg | O(1) | Hash table access |

### Memory Usage Patterns

1. **Dynamic Programming Table**: O(2ⁿ) space for join optimization
2. **Statistics Storage**: O(t × c × h) where t = tables, c = columns, h = histogram buckets  
3. **Plan Cache**: O(p × s) where p = cached plans, s = average plan size
4. **Intermediate Results**: O(n) during rule application

## Performance Optimization Techniques

### 13. Branch-and-Bound Pruning Algorithm

#### Algorithm: `branchAndBoundJoinOpt(relations, current_best_cost)`
```
Input: Relations to join, current best known cost
Output: Optimal join plan or early termination

Algorithm branchAndBoundJoinOpt(relations, current_best_cost):
  1. If |relations| == 1:
     return createScanPlan(relations[0])
  
  2. If |relations| == 2:
     plan = createOptimalJoinPlan(relations[0], relations[1])
     return plan
  
  3. For each valid split (S1, S2) of relations:
     a. Calculate lower bound cost:
        lower_bound = estimateLowerBound(S1, S2)
     
     b. Prune if bound exceeds current best:
        If lower_bound >= current_best_cost:
          continue  // Skip this branch
     
     c. Recursively solve subproblems:
        left_plan = branchAndBoundJoinOpt(S1, current_best_cost)
        right_plan = branchAndBoundJoinOpt(S2, current_best_cost)
     
     d. Create join plan and update best cost:
        join_plan = createJoinPlan(left_plan, right_plan)
        cost = estimateCost(join_plan)
        If cost < current_best_cost:
          current_best_cost = cost
          best_plan = join_plan
  
  4. Return best_plan
```

### 14. Parallel Optimization Algorithm

#### Algorithm: `parallelOptimization(query, worker_count)`
```
Input: Query to optimize, number of worker threads
Output: Optimized plan

Algorithm parallelOptimization(query, worker_count):
  1. Partition optimization work:
     work_items = partitionSearchSpace(query, worker_count)
  
  2. Launch parallel workers:
     results = []
     For each work_item in work_items:
       future = async_execute(optimizePartition, work_item)
       results.append(future)
  
  3. Collect results:
     best_plan = null
     best_cost = infinity
     For each future in results:
       plan = future.get()  // Wait for completion
       cost = estimateCost(plan)
       If cost < best_cost:
         best_cost = cost
         best_plan = plan
  
  4. Return best_plan
```

## Testing and Validation Algorithms

### 15. Plan Correctness Validation Algorithm

#### Algorithm: `validatePlan(physical_plan, logical_plan)`
```
Input: Physical plan, original logical plan
Output: Validation result

Algorithm validatePlan(physical_plan, logical_plan):
  1. Check plan structure:
     a. Verify all logical operators have physical implementations
     b. Check that join conditions are properly handled
     c. Validate sort orders are maintained where required
  
  2. Check data flow:
     a. Verify column references are valid throughout plan
     b. Check that required columns are available at each operator
     c. Validate data types are compatible
  
  3. Check semantic equivalence:
     a. Generate canonical representation of both plans
     b. Compare logical semantics
     c. Verify result schemas match
  
  4. Return validation_result(is_valid, error_messages)
```

This comprehensive set of optimization algorithms provides the foundation for a sophisticated, cost-based query optimizer that can handle complex SQL queries efficiently while maintaining optimal performance characteristics.

## Future Algorithm Enhancements

### 1. Machine Learning-Enhanced Algorithms
- **Neural Network Cost Models**: Replace traditional cost functions
- **Reinforcement Learning Join Ordering**: Learn optimal strategies through experience
- **Cardinality Estimation Networks**: Deep learning for better selectivity prediction

### 2. Approximate Query Processing
- **Sampling-Based Optimization**: Use data samples for faster optimization
- **Progressive Optimization**: Start execution with approximate plans, refine during runtime
- **Multi-Objective Optimization**: Balance multiple goals (cost, latency, resource usage)

### 3. Adaptive Optimization
- **Runtime Plan Adjustment**: Modify plans based on actual execution statistics
- **Feedback-Driven Optimization**: Learn from query execution patterns
- **Workload-Aware Optimization**: Optimize for common query patterns